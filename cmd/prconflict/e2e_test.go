package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v72/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// E2ETestFramework manages the complete end-to-end testing lifecycle
type E2ETestFramework struct {
	ctx      context.Context
	ghREST   *github.Client
	ghQL     *githubv4.Client
	resolver *GraphQLResolver
	token    string
	testRepo *github.Repository
	workDir  string
	owner    string
	repoName string
	cleanup  []func() error
}

// TestScenario defines a complete test scenario
type TestScenario struct {
	Name            string
	InitialFiles    map[string]string // filename -> content
	BranchChanges   map[string]string // files to modify in PR branch
	ReviewComments  []ReviewComment
	ExpectedMarkers []ExpectedMarker
}

// ReviewComment represents a comment to create on the PR
type ReviewComment struct {
	Path    string
	Line    int
	Body    string
	Resolve bool // whether to resolve this comment thread
	ReplyTo int  // if > 0, this is a reply to another comment (by index)
}

// ExpectedMarker represents what we expect to see in the file after prconflict runs
type ExpectedMarker struct {
	File     string
	Line     int
	Contains []string // strings that should be present in the conflict block
}

func NewE2ETestFramework(t *testing.T) *E2ETestFramework {
	t.Helper()

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		t.Skip("GITHUB_TOKEN env var not set")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(ctx, ts)

	ghQL := githubv4.NewClient(httpClient)
	framework := &E2ETestFramework{
		ctx:      ctx,
		ghREST:   github.NewClient(httpClient),
		ghQL:     ghQL,
		resolver: NewGraphQLResolver(ghQL),
		token:    token,
		owner:    "teddyknox", // TODO: make configurable
	}

	return framework
}

func (f *E2ETestFramework) Setup(t *testing.T) error {
	t.Helper()

	// Create unique repo name
	timestamp := time.Now().Unix()
	f.repoName = fmt.Sprintf("prconflict-e2e-test-%d", timestamp)

	// Create temporary working directory
	var err error
	f.workDir, err = os.MkdirTemp("", "prconflict-e2e-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	f.addCleanup(func() error { return os.RemoveAll(f.workDir) })

	// Create GitHub repository
	repo := &github.Repository{
		Name:        github.Ptr(f.repoName),
		Description: github.Ptr("Temporary repo for prconflict e2e testing"),
		Private:     github.Ptr(false),
		AutoInit:    github.Ptr(true),
	}

	f.testRepo, _, err = f.ghREST.Repositories.Create(f.ctx, "", repo)
	if err != nil {
		return fmt.Errorf("failed to create GitHub repo: %w", err)
	}
	f.addCleanup(f.deleteRepo)

	return nil
}

func (f *E2ETestFramework) RunScenario(t *testing.T, scenario TestScenario) error {
	t.Helper()

	t.Logf("Running scenario: %s", scenario.Name)

	// Clone the repository
	if err := f.cloneRepo(); err != nil {
		return fmt.Errorf("clone failed: %w", err)
	}

	// Set up initial files
	if err := f.setupInitialFiles(scenario.InitialFiles); err != nil {
		return fmt.Errorf("initial file setup failed: %w", err)
	}

	// Create and push feature branch with changes
	branchName := "feature-test"
	if err := f.createFeatureBranch(branchName, scenario.BranchChanges); err != nil {
		return fmt.Errorf("feature branch creation failed: %w", err)
	}

	// Create pull request
	pr, err := f.createPR(branchName, "Test PR for e2e testing", "This is a test PR created by the e2e test framework")
	if err != nil {
		return fmt.Errorf("PR creation failed: %w", err)
	}
	t.Logf("Created PR #%d", pr.GetNumber())

	// Add review comments
	if err := f.addReviewComments(pr.GetNumber(), scenario.ReviewComments); err != nil {
		return fmt.Errorf("review comment creation failed: %w", err)
	}

	// Checkout the PR locally
	if err := f.checkoutPR(pr.GetNumber()); err != nil {
		return fmt.Errorf("PR checkout failed: %w", err)
	}

	// Run prconflict tool
	if err := f.runPRConflict(pr.GetNumber()); err != nil {
		return fmt.Errorf("prconflict execution failed: %w", err)
	}

	// Verify results
	if err := f.verifyResults(scenario.ExpectedMarkers); err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	t.Logf("Scenario '%s' completed successfully", scenario.Name)
	return nil
}

func (f *E2ETestFramework) cloneRepo() error {
	repoURL := fmt.Sprintf("https://github.com/%s/%s.git", f.owner, f.repoName)
	cmd := exec.Command("git", "clone", repoURL, f.workDir)
	cmd.Env = append(os.Environ(), fmt.Sprintf("GITHUB_TOKEN=%s", f.token))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w, output: %s", err, output)
	}
	return nil
}

func (f *E2ETestFramework) setupInitialFiles(files map[string]string) error {
	// First, copy go.mod and go.sum from the main project to ensure the test repo is a valid Go module
	mainProjectDir := "../.." // Assuming we're in cmd/prconflict
	goModPath := filepath.Join(mainProjectDir, "go.mod")
	goSumPath := filepath.Join(mainProjectDir, "go.sum")

	// Copy go.mod
	if goModContent, err := os.ReadFile(goModPath); err == nil {
		if err := os.WriteFile(filepath.Join(f.workDir, "go.mod"), goModContent, 0644); err != nil {
			return fmt.Errorf("failed to copy go.mod: %w", err)
		}
	}

	// Copy go.sum if it exists
	if goSumContent, err := os.ReadFile(goSumPath); err == nil {
		if err := os.WriteFile(filepath.Join(f.workDir, "go.sum"), goSumContent, 0644); err != nil {
			return fmt.Errorf("failed to copy go.sum: %w", err)
		}
	}

	// Create cmd/prconflict directory and copy the main.go file
	cmdDir := filepath.Join(f.workDir, "cmd", "prconflict")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		return fmt.Errorf("failed to create cmd/prconflict directory: %w", err)
	}

	// Copy main.go from the current directory
	mainGoPath := "main.go"
	if mainGoContent, err := os.ReadFile(mainGoPath); err == nil {
		if err := os.WriteFile(filepath.Join(cmdDir, "main.go"), mainGoContent, 0644); err != nil {
			return fmt.Errorf("failed to copy main.go: %w", err)
		}
	}

	for filename, content := range files {
		fullPath := filepath.Join(f.workDir, filename)

		// Create directory if needed
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", filename, err)
		}

		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}
	}

	// Commit initial files
	if err := f.gitCommit("Add initial test files"); err != nil {
		return fmt.Errorf("initial commit failed: %w", err)
	}

	// Push to main
	if err := f.gitPush("main"); err != nil {
		return fmt.Errorf("initial push failed: %w", err)
	}

	return nil
}

func (f *E2ETestFramework) createFeatureBranch(branchName string, changes map[string]string) error {
	// Create and checkout branch
	if err := f.gitRun("checkout", "-b", branchName); err != nil {
		return fmt.Errorf("branch creation failed: %w", err)
	}

	// Apply changes
	for filename, content := range changes {
		fullPath := filepath.Join(f.workDir, filename)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}
	}

	// Commit changes
	if err := f.gitCommit("Add changes for PR"); err != nil {
		return fmt.Errorf("branch commit failed: %w", err)
	}

	// Push branch
	if err := f.gitPush(branchName); err != nil {
		return fmt.Errorf("branch push failed: %w", err)
	}

	return nil
}

func (f *E2ETestFramework) createPR(branch, title, body string) (*github.PullRequest, error) {
	pr := &github.NewPullRequest{
		Title: github.Ptr(title),
		Body:  github.Ptr(body),
		Head:  github.Ptr(branch),
		Base:  github.Ptr("main"),
	}

	createdPR, _, err := f.ghREST.PullRequests.Create(f.ctx, f.owner, f.repoName, pr)
	if err != nil {
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	return createdPR, nil
}

func (f *E2ETestFramework) addReviewComments(prNumber int, comments []ReviewComment) error {
	// Get the PR to find the commit SHA
	pr, _, err := f.ghREST.PullRequests.Get(f.ctx, f.owner, f.repoName, prNumber)
	if err != nil {
		return fmt.Errorf("failed to get PR: %w", err)
	}

	commitSHA := pr.GetHead().GetSHA()
	createdComments := make([]*github.PullRequestComment, len(comments))

	for i, comment := range comments {
		if comment.ReplyTo > 0 && comment.ReplyTo <= i {
			// For now, skip reply comments since they have API complexity
			// Just create a regular comment instead
			reviewComment := &github.PullRequestComment{
				Body:     github.Ptr(comment.Body),
				Path:     github.Ptr(comment.Path),
				CommitID: github.Ptr(commitSHA),
				Line:     github.Ptr(comment.Line),
				Side:     github.Ptr("RIGHT"),
			}

			created, _, err := f.ghREST.PullRequests.CreateComment(f.ctx, f.owner, f.repoName, prNumber, reviewComment)
			if err != nil {
				return fmt.Errorf("failed to create review comment: %w", err)
			}
			createdComments[i] = created
		} else {
			// This is a new review comment - need to specify side and use line instead of Line
			reviewComment := &github.PullRequestComment{
				Body:     github.Ptr(comment.Body),
				Path:     github.Ptr(comment.Path),
				CommitID: github.Ptr(commitSHA),
				Line:     github.Ptr(comment.Line),
				Side:     github.Ptr("RIGHT"), // Comments on the new version of the file
			}

			created, _, err := f.ghREST.PullRequests.CreateComment(f.ctx, f.owner, f.repoName, prNumber, reviewComment)
			if err != nil {
				return fmt.Errorf("failed to create review comment: %w", err)
			}
			createdComments[i] = created
		}

		// If this comment should be resolved, resolve it using GraphQL
		if comment.Resolve {
			commentID := createdComments[i].GetID()
			if err := f.resolver.ResolveCommentThread(f.ctx, f.owner, f.repoName, prNumber, commentID); err != nil {
				return fmt.Errorf("failed to resolve comment thread %d: %w", commentID, err)
			}
		}

		// Small delay to ensure proper ordering
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func (f *E2ETestFramework) checkoutPR(prNumber int) error {
	// Fetch the PR branch
	if err := f.gitRun("fetch", "origin", fmt.Sprintf("pull/%d/head:%s", prNumber, fmt.Sprintf("pr-%d", prNumber))); err != nil {
		return fmt.Errorf("failed to fetch PR: %w", err)
	}

	// Checkout the PR branch
	return f.gitRun("checkout", fmt.Sprintf("pr-%d", prNumber))
}

func (f *E2ETestFramework) runPRConflict(prNumber int) error {
	// Build the prconflict binary if needed
	buildCmd := exec.Command("go", "build", "-o", "prconflict", "./cmd/prconflict")
	buildCmd.Dir = f.workDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to build prconflict: %w, output: %s", err, output)
	}

	// Run prconflict
	cmd := exec.Command("./prconflict",
		"--repo", fmt.Sprintf("%s/%s", f.owner, f.repoName),
		"--pr", fmt.Sprintf("%d", prNumber))
	cmd.Dir = f.workDir
	cmd.Env = append(os.Environ(), fmt.Sprintf("GITHUB_TOKEN=%s", f.token))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("prconflict execution failed: %w, output: %s", err, output)
	}

	return nil
}

func (f *E2ETestFramework) verifyResults(expectedMarkers []ExpectedMarker) error {
	for _, marker := range expectedMarkers {
		filePath := filepath.Join(f.workDir, marker.File)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", marker.File, err)
		}

		lines := strings.Split(string(content), "\n")

		// Look for conflict markers around the expected line
		found := false
		searchStart := max(0, marker.Line-5)
		searchEnd := min(len(lines), marker.Line+5)

		for i := searchStart; i < searchEnd; i++ {
			if strings.Contains(lines[i], "<<<<<<< REVIEW THREAD") {
				// Found a conflict block, verify it contains expected content
				blockEnd := i
				for blockEnd < len(lines) && !strings.Contains(lines[blockEnd], ">>>>>>> END REVIEW") {
					blockEnd++
				}

				blockContent := strings.Join(lines[i:blockEnd+1], "\n")
				allFound := true
				for _, expected := range marker.Contains {
					if !strings.Contains(blockContent, expected) {
						allFound = false
						break
					}
				}

				if allFound {
					found = true
					break
				}
			}
		}

		if !found {
			// Enhanced debugging: show the actual file content around the expected line
			debugLines := []string{}
			debugStart := max(0, marker.Line-10)
			debugEnd := min(len(lines), marker.Line+10)
			for i := debugStart; i < debugEnd; i++ {
				prefix := "   "
				if i+1 == marker.Line {
					prefix = ">>>"
				}
				debugLines = append(debugLines, fmt.Sprintf("%s %3d: %s", prefix, i+1, lines[i]))
			}

			return fmt.Errorf("expected conflict marker not found in %s around line %d\nExpected content: %v\nActual file content around line %d:\n%s",
				marker.File, marker.Line, marker.Contains, marker.Line, strings.Join(debugLines, "\n"))
		}
	}

	return nil
}

func (f *E2ETestFramework) gitRun(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = f.workDir
	cmd.Env = append(os.Environ(), fmt.Sprintf("GITHUB_TOKEN=%s", f.token))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s failed: %w, output: %s", strings.Join(args, " "), err, output)
	}
	return nil
}

func (f *E2ETestFramework) gitCommit(message string) error {
	if err := f.gitRun("add", "."); err != nil {
		return err
	}
	return f.gitRun("commit", "-m", message)
}

func (f *E2ETestFramework) gitPush(branch string) error {
	return f.gitRun("push", "origin", branch)
}

func (f *E2ETestFramework) addCleanup(fn func() error) {
	f.cleanup = append(f.cleanup, fn)
}

func (f *E2ETestFramework) deleteRepo() error {
	_, err := f.ghREST.Repositories.Delete(f.ctx, f.owner, f.repoName)
	if err != nil {
		// Don't fail the entire test if repository deletion fails - just log it
		// This can happen if the token doesn't have delete permissions
		return fmt.Errorf("repository deletion failed (non-fatal): %w", err)
	}
	return nil
}

func (f *E2ETestFramework) Cleanup() error {
	var errors []string
	for _, cleanupFn := range f.cleanup {
		if err := cleanupFn(); err != nil {
			// Don't treat repository deletion failures as fatal errors
			if strings.Contains(err.Error(), "repository deletion failed") {
				fmt.Printf("Warning: %s\n", err.Error())
			} else {
				errors = append(errors, err.Error())
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %s", strings.Join(errors, "; "))
	}
	return nil
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
