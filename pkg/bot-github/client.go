package botgithub

import (
	"context"
	"fmt"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// Client wraps the GitHub API client with convenience methods
type Client struct {
	context context.Context
	github  *github.Client
}

// NewClient creates a new GitHub client with the provided token
func NewClient(token string) *Client {
	context := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context, ts)

	return &Client{
		context: context,
		github:  github.NewClient(tc),
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
	mainRef, _, err := client.github.Git.GetRef(client.context, args.Owner, args.Repo, "refs/heads/main")
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

	_, _, err = client.github.Git.CreateRef(client.context, args.Owner, args.Repo, newRef)
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

	_, _, err := client.github.Repositories.CreateFile(
		client.context,
		args.Owner,
		args.Repo,
		args.Filename,
		options,
	)

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

	_, _, err := client.github.Repositories.UpdateFile(
		client.context,
		args.Owner,
		args.Repo,
		args.Filename,
		options,
	)

	if err != nil {
		return fmt.Errorf("updating file: %w", err)
	}

	return nil
}

type DeleteFileArgs struct {
	Branch   string
	Filename string
	Message  string
	Owner    string
	Repo     string
	Sha      string
}

// DeleteFile deletes a file from the repository
func (client *Client) DeleteFile(args DeleteFileArgs) error {
	options := &github.RepositoryContentFileOptions{
		Branch:  github.String(args.Branch),
		Message: github.String(args.Message),
		SHA:     github.String(args.Sha),
	}

	_, _, err := client.github.Repositories.DeleteFile(
		client.context,
		args.Owner,
		args.Repo,
		args.Filename,
		options,
	)

	if err != nil {
		return fmt.Errorf("deleting file: %w", err)
	}

	return nil
}

type CreatePullRequestArgs struct {
	Base  string
	Body  string
	Head  string
	Owner string
	Repo  string
	Title string
}

// CreatePullRequest creates a new pull request
func (client *Client) CreatePullRequest(
	args CreatePullRequestArgs,
) (*github.PullRequest, error) {
	newPR := &github.NewPullRequest{
		Title: github.String(args.Title),
		Head:  github.String(args.Head),
		Base:  github.String(args.Base),
		Body:  github.String(args.Body),
	}

	pullRequest, _, err := client.github.PullRequests.Create(
		client.context,
		args.Owner,
		args.Repo,
		newPR,
	)

	if err != nil {
		return nil, fmt.Errorf("creating PR: %w", err)
	}

	return pullRequest, nil
}

type GetFileContentArgs struct {
	Filename string
	Owner    string
	Ref      string
	Repo     string
}

// GetFileContent retrieves the content of a file from the repository
func (client *Client) GetFileContent(args GetFileContentArgs) (string, string, error) {
	options := &github.RepositoryContentGetOptions{
		Ref: args.Ref,
	}

	fileContent, _, _, err := client.github.Repositories.GetContents(
		client.context,
		args.Owner,
		args.Repo,
		args.Filename,
		options,
	)

	if err != nil {
		return "", "", fmt.Errorf("getting file content: %w", err)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return "", "", fmt.Errorf("decoding content: %w", err)
	}

	return content, *fileContent.SHA, nil
}

type ListPullRequestFilesArgs struct {
	Owner    string
	PrNumber int
	Repo     string
}

// ListPullRequestFiles returns the files changed in a pull request
func (client *Client) ListPullRequestFiles(
	args ListPullRequestFilesArgs,
) ([]*github.CommitFile, error) {
	files, _, err := client.github.PullRequests.ListFiles(
		client.context,
		args.Owner,
		args.Repo,
		args.PrNumber,
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf("listing PR files: %w", err)
	}

	return files, nil
}

type ReactToIssueArgs struct {
	IssueNumber int
	Owner       string
	Reaction    string
	Repo        string
}

// ReactToIssue adds a reaction to an issue
func (client *Client) ReactToIssue(args ReactToIssueArgs) error {
	_, _, err := client.github.Reactions.CreateIssueReaction(
		client.context,
		args.Owner,
		args.Repo,
		args.IssueNumber,
		args.Reaction,
	)

	if err != nil {
		return fmt.Errorf("reacting to issue: %w", err)
	}

	return nil
}

// ReactToPRComment adds a reaction to a PR comment
func (client *Client) ReactToPRComment(owner, repo string, commentID int64, reaction string) error {
	_, _, err := client.github.Reactions.CreatePullRequestCommentReaction(
		client.context,
		owner,
		repo,
		commentID,
		reaction,
	)

	if err != nil {
		return fmt.Errorf("reacting to PR comment: %w", err)
	}

	return nil
}

// CommentOnIssue adds a comment to an issue
func (client *Client) CommentOnIssue(owner, repo string, issueNumber int, comment string) error {
	_, _, err := client.github.Issues.CreateComment(
		client.context,
		owner,
		repo,
		issueNumber,
		&github.IssueComment{
			Body: github.String(comment),
		},
	)

	if err != nil {
		return fmt.Errorf("commenting on issue: %w", err)
	}

	return nil
}

// CommentOnPR adds a comment to a pull request
func (client *Client) CommentOnPR(owner, repo string, prNumber int, comment string) error {
	_, _, err := client.github.Issues.CreateComment(
		client.context,
		owner,
		repo,
		prNumber,
		&github.IssueComment{
			Body: github.String(comment),
		},
	)

	if err != nil {
		return fmt.Errorf("commenting on PR: %w", err)
	}

	return nil
}
