package main

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
)

// GraphQLResolver handles GraphQL operations for thread resolution
type GraphQLResolver struct {
	client *githubv4.Client
}

// ResolveThreadInput represents the input for resolving a thread
type ResolveThreadInput struct {
	ThreadID         string
	ClientMutationID string
}

// ResolveThreadMutation represents the GraphQL mutation to resolve a thread
type ResolveThreadMutation struct {
	ResolveReviewThread struct {
		Thread struct {
			ID         githubv4.String
			IsResolved githubv4.Boolean
		}
		ClientMutationID githubv4.String
	} `graphql:"resolveReviewThread(input: $input)"`
}

// UnresolveThreadMutation represents the GraphQL mutation to unresolve a thread
type UnresolveThreadMutation struct {
	UnresolveReviewThread struct {
		Thread struct {
			ID         githubv4.String
			IsResolved githubv4.Boolean
		}
		ClientMutationID githubv4.String
	} `graphql:"unresolveReviewThread(input: $input)"`
}

// GetThreadIDQuery gets the thread ID for a specific comment
type GetThreadIDQuery struct {
	Repository struct {
		PullRequest struct {
			ReviewThreads struct {
				Nodes []struct {
					ID       githubv4.String
					Comments struct {
						Nodes []struct {
							DatabaseID githubv4.Int
						}
					} `graphql:"comments(first: 10)"`
				}
			} `graphql:"reviewThreads(first: 100)"`
		} `graphql:"pullRequest(number: $pr)"`
	} `graphql:"repository(owner: $owner, name: $name)"`
}

func NewGraphQLResolver(client *githubv4.Client) *GraphQLResolver {
	return &GraphQLResolver{client: client}
}

// GetThreadIDForComment finds the thread ID for a given comment database ID
func (r *GraphQLResolver) GetThreadIDForComment(ctx context.Context, owner, repo string, prNumber int, commentID int64) (string, error) {
	var q GetThreadIDQuery
	vars := map[string]interface{}{
		"owner": githubv4.String(owner),
		"name":  githubv4.String(repo),
		"pr":    githubv4.Int(prNumber),
	}

	if err := r.client.Query(ctx, &q, vars); err != nil {
		return "", fmt.Errorf("failed to query thread ID: %w", err)
	}

	for _, thread := range q.Repository.PullRequest.ReviewThreads.Nodes {
		for _, comment := range thread.Comments.Nodes {
			if int64(comment.DatabaseID) == commentID {
				return string(thread.ID), nil
			}
		}
	}

	return "", fmt.Errorf("thread not found for comment ID %d", commentID)
}

// ResolveThread resolves a review thread by its ID
func (r *GraphQLResolver) ResolveThread(ctx context.Context, threadID string) error {
	var m ResolveThreadMutation
	input := githubv4.ResolveReviewThreadInput{
		ThreadID: githubv4.String(threadID),
	}

	if err := r.client.Mutate(ctx, &m, input, nil); err != nil {
		return fmt.Errorf("failed to resolve thread %s: %w", threadID, err)
	}

	return nil
}

// UnresolveThread unresolves a review thread by its ID
func (r *GraphQLResolver) UnresolveThread(ctx context.Context, threadID string) error {
	var m UnresolveThreadMutation
	input := githubv4.UnresolveReviewThreadInput{
		ThreadID: githubv4.String(threadID),
	}

	if err := r.client.Mutate(ctx, &m, input, nil); err != nil {
		return fmt.Errorf("failed to unresolve thread %s: %w", threadID, err)
	}

	return nil
}

// ResolveCommentThread resolves the thread containing the specified comment
func (r *GraphQLResolver) ResolveCommentThread(ctx context.Context, owner, repo string, prNumber int, commentID int64) error {
	threadID, err := r.GetThreadIDForComment(ctx, owner, repo, prNumber, commentID)
	if err != nil {
		return fmt.Errorf("failed to get thread ID for comment %d: %w", commentID, err)
	}

	return r.ResolveThread(ctx, threadID)
}
