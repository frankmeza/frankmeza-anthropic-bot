// Client implements GitHub API operations that the bot needs.
type Client struct {
	logger *slog.Logger
	client *github.Client
}

// New creates a new GitHub API client.
func New(logger *slog.Logger) *Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	return &Client{
		logger: logger,
		client: client,
	}
}

// ListPullRequestsArgs contains parameters for ListPullRequests method
type ListPullRequestsArgs struct {
	Owner string
	Repo  string
	State string
}

// ListPullRequests gets all pull requests for a repository.
func (client *Client) ListPullRequests(args ListPullRequestsArgs) ([]*github.PullRequest, error) {
	options := &github.PullRequestListOptions{
		State: args.State,
	}

	ctx := context.Background()
	prs, _, err := client.client.PullRequests.List(ctx, args.Owner, args.Repo, options)
	if err != nil {
		return nil, fmt.Errorf("failed to list pull requests: %w", err)
	}
	return prs, nil
}

// GetPullRequestArgs contains parameters for GetPullRequest method
type GetPullRequestArgs struct {
	Owner  string
	Repo   string
	Number int
}

// GetPullRequest gets a specific pull request.
func (client *Client) GetPullRequest(args GetPullRequestArgs) (*github.PullRequest, error) {
	ctx := context.Background()
	pr, _, err := client.client.PullRequests.Get(ctx, args.Owner, args.Repo, args.Number)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request #%d: %w", args.Number, err)
	}
	return pr, nil
}

// GetPullRequestFilesArgs contains parameters for GetPullRequestFiles method
type GetPullRequestFilesArgs struct {
	Owner  string
	Repo   string
	Number int
}

// GetPullRequestFiles gets the files changed in a pull request.
func (client *Client) GetPullRequestFiles(args GetPullRequestFilesArgs) ([]*github.CommitFile, error) {
	ctx := context.Background()
	files, _, err := client.client.PullRequests.ListFiles(ctx, args.Owner, args.Repo, args.Number, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get files for pull request #%d: %w", args.Number, err)
	}
	return files, nil
}

// CommentOnIssueArgs contains parameters for CommentOnIssue method
type CommentOnIssueArgs struct {
	Owner       string
	Repo        string
	IssueNumber int
	Comment     string
}

// CommentOnIssue adds a comment to an issue or pull request.
func (client *Client) CommentOnIssue(args CommentOnIssueArgs) error {
	ctx := context.Background()
	comment := &github.IssueComment{
		Body: github.String(args.Comment),
	}

	_, _, err := client.client.Issues.CreateComment(ctx, args.Owner, args.Repo, args.IssueNumber, comment)
	if err != nil {
		return fmt.Errorf("failed to comment on issue #%d: %w", args.IssueNumber, err)
	}
	return nil
}

// CreatePullRequestReviewArgs contains parameters for CreatePullRequestReview method
type CreatePullRequestReviewArgs struct {
	Owner      string
	Repo       string
	PullNumber int
	Body       string
	Event      string
}

// CreatePullRequestReview creates a review on a pull request.
func (client *Client) CreatePullRequestReview(args CreatePullRequestReviewArgs) error {
	ctx := context.Background()
	review := &github.PullRequestReviewRequest{
		Body:  github.String(args.Body),
		Event: github.String(args.Event),
	}

	_, _, err := client.client.PullRequests.CreateReview(ctx, args.Owner, args.Repo, args.PullNumber, review)
	if err != nil {
		return fmt.Errorf("failed to create review on PR #%d: %w", args.PullNumber, err)
	}
	return nil
}

// GetFileArgs contains parameters for GetFile method
type GetFileArgs struct {
	Owner string
	Repo  string
	Path  string
	Ref   string
}

// GetFile retrieves the content of a file from a repository.
func (client *Client) GetFile(args GetFileArgs) ([]byte, error) {
	ctx := context.Background()
	fileContent, _, _, err := client.client.Repositories.GetContents(
		ctx, args.Owner, args.Repo, args.Path,
		&github.RepositoryContentGetOptions{Ref: args.Ref},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get file %s: %w", args.Path, err)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return nil, fmt.Errorf("failed to decode content of file %s: %w", args.Path, err)
	}

	return []byte(content), nil
}