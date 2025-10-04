package botcode

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	botAi "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot_ai"
	botGithub "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot_github"
	sharedUtils "github.com/frankmeza/frankmeza-anthropic-bot/pkg/shared_utils"
	"github.com/google/go-github/v57/github"
)

// Handler manages webhook events and code operations
type Handler struct {
	AiClient      *botAi.Client
	GithubClient  *botGithub.Client
	Owner         string
	Repo          string
	WebhookSecret string
}

// NewHandler creates a new code handler
func NewHandler(handlerArgs Handler) *Handler {
	return &Handler{
		AiClient:      handlerArgs.AiClient,
		GithubClient:  handlerArgs.GithubClient,
		Owner:         handlerArgs.Owner,
		Repo:          handlerArgs.Repo,
		WebhookSecret: handlerArgs.WebhookSecret,
	}
}

// HandleWebhook processes GitHub webhook events for code changes
func (handler *Handler) HandleWebhook(
	writer http.ResponseWriter,
	request *http.Request,
) {
	payload, err := github.ValidatePayload(request, []byte(handler.WebhookSecret))
	if err != nil {
		log.Printf("webhook validation failed: %v", err)
		http.Error(writer, "validation failed", http.StatusUnauthorized)
		return
	}

	log.Printf("Received payload of length: %d", len(payload))

	event, err := github.ParseWebHook(github.WebHookType(request), payload)
	if err != nil {
		log.Printf("webhook parsing failed: %v", err)
		http.Error(writer, "parsing failed", http.StatusBadRequest)
		return
	}

	switch e := event.(type) {
	case *github.IssuesEvent:
		if *e.Action == "opened" {
			handler.HandleNewIssue(e.Issue)
		}

	case *github.PullRequestReviewCommentEvent:
		if *e.Action == "created" {
			handler.HandlePRComment(e.PullRequest, e.Comment)
		}
	}

	writer.WriteHeader(http.StatusOK)
}

// HandleNewIssue processes new GitHub issues for code changes
func (handler *Handler) HandleNewIssue(issue *github.Issue) {
	title := *issue.Title
	body := *issue.Body

	if !handler.isCodeRequest(title) {
		return
	}

	if err := handler.GithubClient.ReactToIssue(
		botGithub.ReactToIssueArgs{
			Owner:       handler.Owner,
			Repo:        handler.Repo,
			IssueNumber: *issue.Number,
			Reaction:    "+1",
		},
	); err != nil {
		log.Printf("Error reacting to issue: %v", err)
	}

	request := ParseIssueForCodeRequest(title, body)

	if err := handler.createCodeChangePR(issue, request); err != nil {
		log.Printf("Error creating code change PR: %v", err)

		handler.GithubClient.CommentOnIssue(
			botGithub.CommentOnIssueArgs{
				Comment:     "Sorry, I ran into an error creating the code change. Could you check the request format?",
				IssueNumber: *issue.Number,
				Owner:       handler.Owner,
				Repo:        handler.Repo,
			},
		)
	}
}

// createCodeChangePR generates code and creates a PR
func (handler *Handler) createCodeChangePR(
	issue *github.Issue,
	request *ChangeRequest,
) error {
	codeRequest := &botAi.CodeRequest{
		Title:       request.Title,
		Description: request.Description,
		FileType:    request.FileType,
		TargetPath:  request.TargetPath,
		Tags:        request.Tags,
	}

	content, err := handler.AiClient.GenerateCode(codeRequest)
	if err != nil {
		return fmt.Errorf("AI code generation failed: %w", err)
	}

	targetPath := DetermineTargetPath(request)

	codeFile := NewCodeFile(
		CodeFile{
			Content: content,
			Message: GenerateCommitMessage(request, "Add"),
			Path:    targetPath,
		},
	)

	branchName := fmt.Sprintf("ai-code-change-%d", *issue.Number)

	if err := handler.GithubClient.CreateBranch(
		botGithub.CreateBranchArgs{
			BranchName: branchName,
			Owner:      handler.Owner,
			Repo:       handler.Repo,
		},
	); err != nil {
		return fmt.Errorf("creating branch: %w", err)
	}

	message := codeFile.Message

	if err := handler.GithubClient.CreateFile(
		botGithub.CreateFileArgs{
			Branch:   branchName,
			Content:  codeFile.Content,
			Filename: codeFile.Path,
			Message:  message,
			Owner:    handler.Owner,
			Repo:     handler.Repo,
		},
	); err != nil {
		return fmt.Errorf("creating file: %w", err)
	}

	title := fmt.Sprintf("Add code: %s", request.Title)
	body := handler.generatePRBody(issue, codeFile)
	head := fmt.Sprintf("%s:%s", handler.Owner, branchName)

	_, err = handler.GithubClient.CreatePullRequest(
		botGithub.CreatePullRequestArgs{
			Base:  "main",
			Body:  body,
			Head:  head,
			Owner: handler.Owner,
			Repo:  handler.Repo,
			Title: title,
		},
	)

	if err != nil {
		return fmt.Errorf("creating PR: %w", err)
	}

	return nil
}

// HandlePRComment processes comments on pull requests
func (handler *Handler) HandlePRComment(
	pullRequest *github.PullRequest,
	comment *github.PullRequestComment,
) {
	commentBody := *comment.Body

	if err := handler.GithubClient.ReactToPRComment(
		botGithub.ReactToPRCommentArgs{
			Owner:     handler.Owner,
			Repo:      handler.Repo,
			CommentID: *comment.ID,
			Reaction:  "+1",
		},
	); err != nil {
		log.Printf("Error reacting to PR comment: %v", err)
	}

	if !handler.isChangeRequest(commentBody) {
		return
	}

	if err := handler.handleCodeModification(pullRequest, commentBody); err != nil {
		log.Printf("Error updating code: %v", err)

		handler.GithubClient.CommentOnPR(
			botGithub.CommentOnPRArgs{
				Comment:  "Sorry, I had trouble making that change. Could you be more specific?",
				Owner:    handler.Owner,
				PrNumber: *pullRequest.Number,
				Repo:     handler.Repo,
			},
		)

		return
	}

	handler.GithubClient.ReactToPRComment(
		botGithub.ReactToPRCommentArgs{
			Owner:     handler.Owner,
			Repo:      handler.Repo,
			CommentID: *comment.ID,
			Reaction:  "rocket",
		},
	)
}

// handleCodeModification modifies code based on feedback
func (handler *Handler) handleCodeModification(
	pullRequest *github.PullRequest,
	changeRequest string,
) error {
	files, err := handler.GithubClient.ListPullRequestFiles(
		botGithub.ListPullRequestFilesArgs{
			Owner:    handler.Owner,
			Repo:     handler.Repo,
			PrNumber: *pullRequest.Number,
		},
	)

	if err != nil {
		return fmt.Errorf("getting PR files: %w", err)
	}

	for _, file := range files {
		if !strings.HasSuffix(*file.Filename, ".go") {
			continue
		}

		currentContent, sha, err := handler.GithubClient.GetFileContent(
			botGithub.GetFileContentArgs{
				Filename: *file.Filename,
				Owner:    handler.Owner,
				Ref:      *pullRequest.Head.Ref,
				Repo:     handler.Repo,
			},
		)

		if err != nil {
			return fmt.Errorf("getting file content: %w", err)
		}

		updatedContent, err := handler.AiClient.ModifyCode(
			currentContent,
			changeRequest,
		)

		if err != nil {
			return fmt.Errorf("AI modification failed: %w", err)
		}

		message := fmt.Sprintf(
			"Update code based on feedback: %s",
			sharedUtils.TruncateText(changeRequest, 50),
		)

		if err := handler.GithubClient.UpdateFile(
			botGithub.UpdateFileArgs{
				Branch:   *pullRequest.Head.Ref,
				Content:  updatedContent,
				Filename: *file.Filename,
				Message:  message,
				Owner:    handler.Owner,
				Repo:     handler.Repo,
				Sha:      sha,
			},
		); err != nil {
			return fmt.Errorf("updating file: %w", err)
		}

		break
	}

	return nil
}

// Helper methods

func (handler *Handler) isCodeRequest(title string) bool {
	lowerTitle := strings.ToLower(title)

	return strings.Contains(lowerTitle, "code:") ||
		strings.Contains(lowerTitle, "add feature") ||
		strings.Contains(lowerTitle, "refactor") ||
		strings.Contains(lowerTitle, "implement")
}

func (handler *Handler) isChangeRequest(comment string) bool {
	changeWords := []string{
		"can you", "could you", "please", "add", "remove", "change", "update",
		"make it", "make this", "more", "less", "fix", "improve", "rewrite", "refactor",
	}

	lowerComment := strings.ToLower(comment)

	for _, word := range changeWords {
		if strings.Contains(lowerComment, word) {
			return true
		}
	}

	return false
}

func (handler *Handler) generatePRBody(issue *github.Issue, codeFile *CodeFile) string {
	return fmt.Sprintf(`🤖 AI-generated code change based on issue #%d

**File:** %s
**Description:** %s

This code was automatically generated. Feel free to comment with any changes you'd like me to make!

Closes #%d`, *issue.Number, codeFile.Path, *issue.Title, *issue.Number)
}
