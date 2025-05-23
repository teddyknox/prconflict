package main

import (
	"context"
	"os"
	"fmt"
	"testing"

	"github.com/google/go-github/v72/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

var (
	testOwner string = "teddyknox"
	testRepo  string = "prconflict"
	testPR    int    = 1
)

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
	fmt.Printf("Fetched %d unresolved comment IDs\n", len(ids))
	for id := range ids {
		fmt.Printf("Unresolved ID: %d\n", id)
	}
	// Validate expected IDs
	expectedIDs := []int64{2104860587, 2104861653}
	if len(ids) != len(expectedIDs) {
		t.Fatalf("expected %d unresolved IDs, got %d", len(expectedIDs), len(ids))
	}
	for _, e := range expectedIDs {
		if _, ok := ids[e]; !ok {
			t.Errorf("expected unresolved comment ID %d", e)
		}
	}
	// Ensure map is non-nil
	if ids == nil {
		t.Fatalf("expected non-nil map of IDs")
	}
}

func TestIntegration_FetchReviewComments(t *testing.T) {
	ctx, ghREST, _, owner, repo, prNumber := setupClients(t)
	comments := fetchReviewComments(ctx, ghREST, owner, repo, prNumber)
	fmt.Printf("Fetched %d total review comments\n", len(comments))
	for _, c := range comments {
		fmt.Printf("Comment ID: %d, Path: %s, Line: %d\n", c.GetID(), c.GetPath(), c.GetLine())
	}
	// Validate expected comments
	expected := []struct{ID int64; Path string; Line int}{
		{2104860587, "cmd/prconflict/integration_test.go", 17},
		{2104861133, "cmd/prconflict/integration_test.go", 34},
		{2104861653, "cmd/prconflict/integration_test.go", 17},
	}
	if len(comments) != len(expected) {
		t.Fatalf("expected %d comments, got %d", len(expected), len(comments))
	}
	for i, exp := range expected {
		c := comments[i]
		if c.GetID() != exp.ID || c.GetPath() != exp.Path || c.GetLine() != exp.Line {
			t.Errorf("comment[%d] = (%d,%q,%d); want (%d,%q,%d)", i, c.GetID(), c.GetPath(), c.GetLine(), exp.ID, exp.Path, exp.Line)
		}
	}
	if comments == nil {
		t.Fatalf("expected non-nil slice of comments")
	}
}

func TestIntegration_ConsistencyBetweenGraphQLAndREST(t *testing.T) {
	ctx, ghREST, ghQL, owner, repo, prNumber := setupClients(t)
	ids := getUnresolvedCommentIDs(ctx, ghQL, owner, repo, prNumber)
	fmt.Printf("Unresolved IDs count: %d\n", len(ids))
	for id := range ids {
		fmt.Printf("Unresolved ID: %d\n", id)
	}
	comments := fetchReviewComments(ctx, ghREST, owner, repo, prNumber)
	fmt.Printf("Total comments fetched: %d\n", len(comments))
	for _, c := range comments {
		fmt.Printf("Fetched comment ID: %d\n", c.GetID())
	}

	if len(ids) != 2 {
		t.Fatalf("expected 2 unresolved IDs, got %d", len(ids))
	}
	if len(comments) != 3 {
		t.Fatalf("expected 3 total comments, got %d", len(comments))
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
