package botgithub

import (
	"context"
	"fmt"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// Client wraps the GitHub API client with convenience methods
type Client struct {
	github *github.Client
	ctx    context.Context
}

// NewClient creates a new GitHub client with the provided token
func NewClient(token string) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		github: github.NewClient(tc),
		ctx:    ctx,
	}
}

// CreateBranch creates a new branch from the main branch
func (c *Client) CreateBranch(owner, repo, branchName string) error {
	// Get the main branch reference
	mainRef, _, err := c.github.Git.GetRef(c.ctx, owner, repo, "refs/heads/main")
	if err != nil {
		return fmt.Errorf("getting main branch: %w", err)
	}

	// Create new branch reference
	newRef := &github.Reference{
		Ref: github.String("refs/heads/" + branchName),
		Object: &github.GitObject{
			SHA: mainRef.Object.SHA,
		},
	}

	_, _, err = c.github.Git.CreateRef(c.ctx, owner, repo, newRef)
	if err != nil {
		return fmt.Errorf("creating branch: %w", err)
	}

	return nil
}

// CreateFile creates a new file in the repository
func (c *Client) CreateFile(owner, repo, branch, filename, content, message string) error {
	opts := &github.RepositoryContentFileOptions{
		Message: github.String(message),
		Content: []byte(content),
		Branch:  github.String(branch),
	}

	_, _, err := c.github.Repositories.CreateFile(c.ctx, owner, repo, filename, opts)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}

	return nil
}

// UpdateFile updates an existing file in the repository
func (c *Client) UpdateFile(owner, repo, branch, filename, content, message, sha string) error {
	opts := &github.RepositoryContentFileOptions{
		Message: github.String(message),
		Content: []byte(content),
		Branch:  github.String(branch),
		SHA:     github.String(sha),
	}

	_, _, err := c.github.Repositories.UpdateFile(c.ctx, owner, repo, filename, opts)
	if err != nil {
		return fmt.Errorf("updating file: %w", err)
	}

	return nil
}

// DeleteFile deletes a file from the repository
func (c *Client) DeleteFile(owner, repo, branch, filename, message, sha string) error {
	opts := &github.RepositoryContentFileOptions{
		Message: github.String(message),
		Branch:  github.String(branch),
		SHA:     github.String(sha),
	}

	_, _, err := c.github.Repositories.DeleteFile(c.ctx, owner, repo, filename, opts)
	if err != nil {
		return fmt.Errorf("deleting file: %w", err)
	}

	return nil
}

// CreatePullRequest creates a new pull request
func (c *Client) CreatePullRequest(owner, repo, title, body, head, base string) (*github.PullRequest, error) {
	newPR := &github.NewPullRequest{
		Title: github.String(title),
		Head:  github.String(head),
		Base:  github.String(base),
		Body:  github.String(body),
	}

	pr, _, err := c.github.PullRequests.Create(c.ctx, owner, repo, newPR)
	if err != nil {
		return nil, fmt.Errorf("creating PR: %w", err)
	}

	return pr, nil
}

// GetFileContent retrieves the content of a file from the repository
func (c *Client) GetFileContent(owner, repo, filename, ref string) (string, string, error) {
	opts := &github.RepositoryContentGetOptions{
		Ref: ref,
	}

	fileContent, _, _, err := c.github.Repositories.GetContents(c.ctx, owner, repo, filename, opts)
	if err != nil {
		return "", "", fmt.Errorf("getting file content: %w", err)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return "", "", fmt.Errorf("decoding content: %w", err)
	}

	return content, *fileContent.SHA, nil
}

// ListPullRequestFiles returns the files changed in a pull request
func (c *Client) ListPullRequestFiles(owner, repo string, prNumber int) ([]*github.CommitFile, error) {
	files, _, err := c.github.PullRequests.ListFiles(c.ctx, owner, repo, prNumber, nil)
	if err != nil {
		return nil, fmt.Errorf("listing PR files: %w", err)
	}

	return files, nil
}

// ReactToIssue adds a reaction to an issue
func (c *Client) ReactToIssue(owner, repo string, issueNumber int, reaction string) error {
	_, _, err := c.github.Reactions.CreateIssueReaction(c.ctx, owner, repo, issueNumber, reaction)

	if err != nil {
		return fmt.Errorf("reacting to issue: %w", err)
	}

	return nil
}

// ReactToPRComment adds a reaction to a PR comment
func (c *Client) ReactToPRComment(owner, repo string, commentID int64, reaction string) error {
	_, _, err := c.github.Reactions.CreatePullRequestCommentReaction(c.ctx, owner, repo, commentID, reaction)

	if err != nil {
		return fmt.Errorf("reacting to PR comment: %w", err)
	}

	return nil
}

// CommentOnIssue adds a comment to an issue
func (c *Client) CommentOnIssue(owner, repo string, issueNumber int, comment string) error {
	_, _, err := c.github.Issues.CreateComment(c.ctx, owner, repo, issueNumber, &github.IssueComment{
		Body: github.String(comment),
	})

	if err != nil {
		return fmt.Errorf("commenting on issue: %w", err)
	}

	return nil
}

// CommentOnPR adds a comment to a pull request
func (c *Client) CommentOnPR(owner, repo string, prNumber int, comment string) error {
	_, _, err := c.github.Issues.CreateComment(c.ctx, owner, repo, prNumber, &github.IssueComment{
		Body: github.String(comment),
	})

	if err != nil {
		return fmt.Errorf("commenting on PR: %w", err)
	}

	return nil
}
