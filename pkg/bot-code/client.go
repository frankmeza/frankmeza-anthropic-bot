// client.go
package botgithub

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type Client struct {
	client      *github.Client
	rateLimiter <-chan time.Time
}

func NewClient(token string, requestsPerHour int) (*Client, error) {
	if token == "" {
		return nil, errors.New("GitHub token is required")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Create a rate limiter that allows `requestsPerHour` requests per hour
	// This is to avoid hitting GitHub's rate limit
	var rateLimiter <-chan time.Time
	if requestsPerHour > 0 {
		rateLimiter = time.Tick(time.Hour / time.Duration(requestsPerHour))
	}

	return &Client{
		client:      client,
		rateLimiter: rateLimiter,
	}, nil
}

// waitForRateLimit waits for the rate limiter if it's set
func (client *Client) waitForRateLimit() {
	if client.rateLimiter != nil {
		<-client.rateLimiter
	}
}

type GetIssueArgs struct {
	Owner       string
	Repo        string
	IssueNumber int
}

// GetIssue retrieves an issue from a GitHub repository
func (client *Client) GetIssue(args GetIssueArgs) (*github.Issue, error) {
	client.waitForRateLimit()

	issue, _, err := client.client.Issues.Get(context.Background(), args.Owner, args.Repo, args.IssueNumber)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get GitHub issue")
	}

	return issue, nil
}

type GetIssueCommentsArgs struct {
	Owner       string
	Repo        string
	IssueNumber int
}

// GetIssueComments retrieves all comments for an issue
func (client *Client) GetIssueComments(args GetIssueCommentsArgs) ([]*github.IssueComment, error) {
	client.waitForRateLimit()

	var allComments []*github.IssueComment
	listOptions := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{
			PerPage: 100, // Max allowed by GitHub API
		},
	}

	for {
		comments, resp, err := client.client.Issues.ListComments(
			context.Background(),
			args.Owner,
			args.Repo,
			args.IssueNumber,
			listOptions,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to list GitHub issue comments")
		}

		allComments = append(allComments, comments...)

		if resp.NextPage == 0 {
			break
		}

		listOptions.Page = resp.NextPage
		client.waitForRateLimit()
	}

	return allComments, nil
}

type CommentOnIssueArgs struct {
	Owner       string
	Repo        string
	IssueNumber int
	Comment     string
}

// CommentOnIssue adds a new comment to a GitHub issue
func (client *Client) CommentOnIssue(args CommentOnIssueArgs) error {
	client.waitForRateLimit()

	comment := &github.IssueComment{
		Body: github.String(args.Comment),
	}

	_, _, err := client.client.Issues.CreateComment(
		context.Background(),
		args.Owner,
		args.Repo,
		args.IssueNumber,
		comment,
	)
	if err != nil {
		return errors.Wrap(err, "failed to create GitHub issue comment")
	}

	return nil
}

type CloseIssueArgs struct {
	Owner       string
	Repo        string
	IssueNumber int
}

// CloseIssue closes a GitHub issue
func (client *Client) CloseIssue(args CloseIssueArgs) error {
	client.waitForRateLimit()

	issueRequest := &github.IssueRequest{
		State: github.String("closed"),
	}

	_, _, err := client.client.Issues.Edit(
		context.Background(),
		args.Owner,
		args.Repo,
		args.IssueNumber,
		issueRequest,
	)
	if err != nil {
		return errors.Wrap(err, "failed to close GitHub issue")
	}

	return nil
}

type AddLabelsToIssueArgs struct {
	Owner       string
	Repo        string
	IssueNumber int
	Labels      []string
}

// AddLabelsToIssue adds labels to a GitHub issue
func (client *Client) AddLabelsToIssue(args AddLabelsToIssueArgs) error {
	client.waitForRateLimit()

	_, _, err := client.client.Issues.AddLabelsToIssue(
		context.Background(),
		args.Owner,
		args.Repo,
		args.IssueNumber,
		args.Labels,
	)
	if err != nil {
		return errors.Wrap(err, "failed to add labels to GitHub issue")
	}

	return nil
}

type GetUserPublicEmailsArgs struct {
	Username string
}

// GetUserPublicEmails attempts to find public email addresses for a GitHub user
func (client *Client) GetUserPublicEmails(args GetUserPublicEmailsArgs) ([]string, error) {
	client.waitForRateLimit()

	// First, get the user's profile
	user, _, err := client.client.Users.Get(context.Background(), args.Username)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get GitHub user")
	}

	// Check if the user has a public email
	var emails []string
	if user.Email != nil && *user.Email != "" {
		emails = append(emails, *user.Email)
	}

	return emails, nil
}

type GetRepositoryMetadataArgs struct {
	Owner string
	Repo  string
}

// GetRepositoryMetadata retrieves metadata about a GitHub repository
func (client *Client) GetRepositoryMetadata(args GetRepositoryMetadataArgs) (*github.Repository, error) {
	client.waitForRateLimit()

	repo, _, err := client.client.Repositories.Get(context.Background(), args.Owner, args.Repo)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to get repository %s/%s", args.Owner, args.Repo))
	}

	return repo, nil
}

type GetPullRequestArgs struct {
	Owner  string
	Repo   string
	Number int
}

// GetPullRequest retrieves a pull request from a GitHub repository
func (client *Client) GetPullRequest(args GetPullRequestArgs) (*github.PullRequest, error) {
	client.waitForRateLimit()

	pr, _, err := client.client.PullRequests.Get(context.Background(), args.Owner, args.Repo, args.Number)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get GitHub pull request")
	}

	return pr, nil
}

type GetPullRequestFilesArgs struct {
	Owner  string
	Repo   string
	Number int
}

// GetPullRequestFiles retrieves the list of files changed in a pull request
func (client *Client) GetPullRequestFiles(args GetPullRequestFilesArgs) ([]*github.CommitFile, error) {
	client.waitForRateLimit()

	var allFiles []*github.CommitFile
	listOptions := &github.ListOptions{
		PerPage: 100, // Max allowed by GitHub API
	}

	for {
		files, resp, err := client.client.PullRequests.ListFiles(
			context.Background(),
			args.Owner,
			args.Repo,
			args.Number,
			listOptions,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to list GitHub pull request files")
		}

		allFiles = append(allFiles, files...)

		if resp.NextPage == 0 {
			break
		}

		listOptions.Page = resp.NextPage
		client.waitForRateLimit()
	}

	return allFiles, nil
}