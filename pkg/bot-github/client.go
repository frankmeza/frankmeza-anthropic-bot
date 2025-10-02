package botgithub

import (
	"context"
	"fmt"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// Client wraps the GitHub API client with convenience methods
type Client struct {
	ctx    context.Context
	github *github.Client
}

// NewClient creates a new GitHub client with the provided token
func NewClient(token string) *Client {
	context := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context, ts)

	return &Client{
		ctx:    context,
		github: github.NewClient(tc),
	}
}

type CreateBranchArgs struct {
	BranchName string
	Owner      string
	Repo       string
}

// CreateBranch creates a new branch from the main branch
func (client *Client) CreateBranch(args CreateBranchArgs) error {
	// Get the main branch reference
	mainRef, _, err := client.github.Git.GetRef(client.ctx, args.Owner, args.Repo, "refs/heads/main")
	if err != nil {
		return fmt.Errorf("getting main branch: %w", err)
	}

	// Create new branch reference
	newRef := &github.Reference{
		Object: &github.GitObject{
			SHA: mainRef.Object.SHA,
		},
		Ref: github.String("refs/heads/" + args.BranchName),
	}

	_, _, err = client.github.Git.CreateRef(client.ctx, args.Owner, args.Repo, newRef)
	if err != nil {
		return fmt.Errorf("creating branch: %w", err)
	}

	return nil
}

type CreateFileArgs struct {
	Branch   string
	Content  string
	Filename string
	Message  string
	Owner    string
	Repo     string
}

// CreateFile creates a new file in the repository
func (client *Client) CreateFile(args CreateFileArgs) error {
	options := &github.RepositoryContentFileOptions{
		Message: github.String(args.Message),
		Content: []byte(args.Content),
		Branch:  github.String(args.Branch),
	}

	_, _, err := client.github.Repositories.CreateFile(client.ctx, args.Owner, args.Repo, args.Filename, options)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}

	return nil
}

type UpdateFileArgs struct {
	Branch   string
	Content  string
	Filename string
	Message  string
	Owner    string
	Repo     string
	Sha      string
}

// UpdateFile updates an existing file in the repository
func (client *Client) UpdateFile(args UpdateFileArgs) error {
	options := &github.RepositoryContentFileOptions{
		Branch:  github.String(args.Branch),
		Content: []byte(args.Content),
		Message: github.String(args.Message),
		SHA:     github.String(args.Sha),
	}

	_, _, err := client.github.Repositories.UpdateFile(client.ctx, args.Owner, args.Repo, args.Filename, options)
	if err != nil {
		return fmt.Errorf("updating file: %w", err)
	}

	return nil
}

// DeleteFile deletes a file from the repository
func (client *Client) DeleteFile(owner, repo, branch, filename, message, sha string) error {
	options := &github.RepositoryContentFileOptions{
		Branch:  github.String(branch),
		Message: github.String(message),
		SHA:     github.String(sha),
	}

	_, _, err := client.github.Repositories.DeleteFile(client.ctx, owner, repo, filename, options)
	if err != nil {
		return fmt.Errorf("deleting file: %w", err)
	}

	return nil
}

// CreatePullRequest creates a new pull request
func (client *Client) CreatePullRequest(owner, repo, title, body, head, base string) (*github.PullRequest, error) {
	newPR := &github.NewPullRequest{
		Title: github.String(title),
		Head:  github.String(head),
		Base:  github.String(base),
		Body:  github.String(body),
	}

	pullRequest, _, err := client.github.PullRequests.Create(client.ctx, owner, repo, newPR)
	if err != nil {
		return nil, fmt.Errorf("creating PR: %w", err)
	}

	return pullRequest, nil
}

// GetFileContent retrieves the content of a file from the repository
func (client *Client) GetFileContent(owner, repo, filename, ref string) (string, string, error) {
	options := &github.RepositoryContentGetOptions{
		Ref: ref,
	}

	fileContent, _, _, err := client.github.Repositories.GetContents(client.ctx, owner, repo, filename, options)
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
func (client *Client) ListPullRequestFiles(owner, repo string, prNumber int) ([]*github.CommitFile, error) {
	files, _, err := client.github.PullRequests.ListFiles(client.ctx, owner, repo, prNumber, nil)
	if err != nil {
		return nil, fmt.Errorf("listing PR files: %w", err)
	}

	return files, nil
}

// ReactToIssue adds a reaction to an issue
func (client *Client) ReactToIssue(owner, repo string, issueNumber int, reaction string) error {
	_, _, err := client.github.Reactions.CreateIssueReaction(client.ctx, owner, repo, issueNumber, reaction)

	if err != nil {
		return fmt.Errorf("reacting to issue: %w", err)
	}

	return nil
}

// ReactToPRComment adds a reaction to a PR comment
func (client *Client) ReactToPRComment(owner, repo string, commentID int64, reaction string) error {
	_, _, err := client.github.Reactions.CreatePullRequestCommentReaction(client.ctx, owner, repo, commentID, reaction)

	if err != nil {
		return fmt.Errorf("reacting to PR comment: %w", err)
	}

	return nil
}

// CommentOnIssue adds a comment to an issue
func (client *Client) CommentOnIssue(owner, repo string, issueNumber int, comment string) error {
	_, _, err := client.github.Issues.CreateComment(client.ctx, owner, repo, issueNumber, &github.IssueComment{
		Body: github.String(comment),
	})

	if err != nil {
		return fmt.Errorf("commenting on issue: %w", err)
	}

	return nil
}

// CommentOnPR adds a comment to a pull request
func (client *Client) CommentOnPR(owner, repo string, prNumber int, comment string) error {
	_, _, err := client.github.Issues.CreateComment(client.ctx, owner, repo, prNumber, &github.IssueComment{
		Body: github.String(comment),
	})

	if err != nil {
		return fmt.Errorf("commenting on PR: %w", err)
	}

	return nil
}
