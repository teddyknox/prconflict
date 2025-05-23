package main

import (
	"context"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-github/v72/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

var (
	testOwner string = "teddyknox"
	testRepo  string = "prconflict"
	testPR    int    = "kkkkkkkkkk"
)

func init() {
	repoEnv := os.Getenv("GITHUB_REPOSITORY")
	parts := strings.SplitN(repoEnv, "/", 2)
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		testOwner, testRepo = parts[0], parts[1]
	}
	if prEnv := os.Getenv("GITHUB_PR_NUMBER"); prEnv != "" {
		if n, err := strconv.Atoi(prEnv); err == nil {
			testPR = n
		}
	}
}

func setupClients(t *testing.T) (ctx context.Context, ghREST *github.Client, ghQL *githubv4.Client, owner, repo string, prNumber int) {
	t.Helper()
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		t.Skip("GITHUB_TOKEN env var not set")
	}

	if testOwner == "" || testRepo == "" {
		t.Skip("test owner/repo not set")
	}
	owner, repo = testOwner, testRepo

	if testPR == 0 {
		t.Skip("test PR number not set")
	}
	prNumber = testPR

	ctx = context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(ctx, ts)
	ghREST = github.NewClient(httpClient)
	ghQL = githubv4.NewClient(httpClient)
	return
}

func TestIntegration_GetUnresolvedCommentIDs(t *testing.T) {
	ctx, _, ghQL, owner, repo, prNumber := setupClients(t)
	ids := getUnresolvedCommentIDs(ctx, ghQL, owner, repo, prNumber)
	t.Logf("Fetched %d unresolved comment IDs", len(ids))
	// Ensure map is non-nil
	if ids == nil {
		t.Fatalf("expected non-nil map of IDs")
	}
}

func TestIntegration_FetchReviewComments(t *testing.T) {
	ctx, ghREST, _, owner, repo, prNumber := setupClients(t)
	comments := fetchReviewComments(ctx, ghREST, owner, repo, prNumber)
	t.Logf("Fetched %d total review comments", len(comments))
	if comments == nil {
		t.Fatalf("expected non-nil slice of comments")
	}
}

func TestIntegration_ConsistencyBetweenGraphQLAndREST(t *testing.T) {
	ctx, ghREST, ghQL, owner, repo, prNumber := setupClients(t)
	ids := getUnresolvedCommentIDs(ctx, ghQL, owner, repo, prNumber)
	comments := fetchReviewComments(ctx, ghREST, owner, repo, prNumber)

	if len(ids) == 0 {
		t.Skip("no unresolved comment IDs found; skipping consistency check")
	}

	// Build set of IDs from fetched comments
	found := make(map[int64]struct{})
	for _, c := range comments {
		found[c.GetID()] = struct{}{}
	}

	// Ensure every unresolved ID is present in fetched comments
	for id := range ids {
		if _, ok := found[id]; !ok {
			t.Errorf("unresolved comment ID %d not found in REST comments", id)
		}
	}
}
