package botcode

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	botai "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot-ai"
	botgithub "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot-github"
	"github.com/google/go-github/v57/github"
)

// Handler manages webhook events and code operations
type Handler struct {
	githubClient  *botgithub.Client
	aiClient      *botai.Client
	owner         string
	repo          string
	webhookSecret string
}

// NewHandler creates a new code handler
func NewHandler(githubClient *botgithub.Client, aiClient *botai.Client, owner, repo, webhookSecret string) *Handler {
	return &Handler{
		githubClient:  githubClient,
		aiClient:      aiClient,
		owner:         owner,
		repo:          repo,
		webhookSecret: webhookSecret,
	}
}

// HandleWebhook processes GitHub webhook events for code changes
func (handler *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// payload, err := github.ValidatePayload(r, []byte(""))
	// // payload, err := github.ValidatePayload(r, []byte(handler.webhookSecret))
	// if err != nil {
	// 	log.Printf("webhook validation failed: %v", err)
	// 	http.Error(w, "validation failed", http.StatusUnauthorized)
	// 	return
	// }

	// // Temporarily skip validation for debugging
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "error reading body", http.StatusBadRequest)
		return
	}

	log.Printf("Received payload of length: %d", len(payload))

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		log.Printf("webhook parsing failed: %v", err)
		http.Error(w, "parsing failed", http.StatusBadRequest)
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

	w.WriteHeader(http.StatusOK)
}

// HandleNewIssue processes new GitHub issues for code changes
func (handler *Handler) HandleNewIssue(issue *github.Issue) {
	title := *issue.Title
	body := *issue.Body

	if !handler.isCodeRequest(title) {
		return
	}

	if err := handler.githubClient.ReactToIssue(
		botgithub.ReactToIssueArgs{
			Owner:       handler.owner,
			Repo:        handler.repo,
			IssueNumber: *issue.Number,
			Reaction:    "+1",
		},
	); err != nil {
		log.Printf("Error reacting to issue: %v", err)
	}

	request := ParseIssueForCodeRequest(title, body)

	if err := handler.createCodeChangePR(issue, request); err != nil {
		log.Printf("Error creating code change PR: %v", err)

		handler.githubClient.CommentOnIssue(handler.owner, handler.repo, *issue.Number,
			"Sorry, I ran into an error creating the code change. Could you check the request format?")
	}
}

// createCodeChangePR generates code and creates a PR
func (handler *Handler) createCodeChangePR(issue *github.Issue, request *ChangeRequest) error {
	codeRequest := &botai.CodeRequest{
		Title:       request.Title,
		Description: request.Description,
		FileType:    request.FileType,
		TargetPath:  request.TargetPath,
		Tags:        request.Tags,
	}

	content, err := handler.aiClient.GenerateCode(codeRequest)
	if err != nil {
		return fmt.Errorf("AI code generation failed: %w", err)
	}

	targetPath := DetermineTargetPath(request)
	codeFile := NewCodeFile(targetPath, content, GenerateCommitMessage(request, "Add"))

	branchName := fmt.Sprintf("ai-code-change-%d", *issue.Number)

	if err := handler.githubClient.CreateBranch(
		botgithub.CreateBranchArgs{
			BranchName: branchName,
			Owner:      handler.owner,
			Repo:       handler.repo,
		},
	); err != nil {
		return fmt.Errorf("creating branch: %w", err)
	}

	message := codeFile.Message

	if err := handler.githubClient.CreateFile(
		botgithub.CreateFileArgs{
			Branch:   branchName,
			Content:  codeFile.Content,
			Filename: codeFile.Path,
			Message:  message,
			Owner:    handler.owner,
			Repo:     handler.repo,
		},
	); err != nil {
		return fmt.Errorf("creating file: %w", err)
	}

	title := fmt.Sprintf("Add code: %s", request.Title)
	body := handler.generatePRBody(issue, codeFile)
	head := fmt.Sprintf("%s:%s", handler.owner, branchName)

	_, err = handler.githubClient.CreatePullRequest(
		botgithub.CreatePullRequestArgs{
			Owner: handler.owner,
			Repo:  handler.repo,
			Title: title,
			Body:  body,
			Head:  head,
			Base:  "main",
		},
	)
	if err != nil {
		return fmt.Errorf("creating PR: %w", err)
	}

	return nil
}

// HandlePRComment processes comments on pull requests
func (handler *Handler) HandlePRComment(pr *github.PullRequest, comment *github.PullRequestComment) {
	commentBody := *comment.Body

	if err := handler.githubClient.ReactToPRComment(handler.owner, handler.repo, *comment.ID, "+1"); err != nil {
		log.Printf("Error reacting to PR comment: %v", err)
	}

	if !handler.isChangeRequest(commentBody) {
		return
	}

	if err := handler.handleCodeModification(pr, commentBody); err != nil {
		log.Printf("Error updating code: %v", err)

		handler.githubClient.CommentOnPR(handler.owner, handler.repo, *pr.Number,
			"Sorry, I had trouble making that change. Could you be more specific?")

		return
	}

	handler.githubClient.ReactToPRComment(handler.owner, handler.repo, *comment.ID, "🚀")
}

// handleCodeModification modifies code based on feedback
func (handler *Handler) handleCodeModification(
	pr *github.PullRequest,
	changeRequest string,
) error {
	files, err := handler.githubClient.ListPullRequestFiles(
		botgithub.ListPullRequestFilesArgs{
			Owner:    handler.owner,
			Repo:     handler.repo,
			PrNumber: *pr.Number,
		},
	)

	if err != nil {
		return fmt.Errorf("getting PR files: %w", err)
	}

	for _, file := range files {
		if !strings.HasSuffix(*file.Filename, ".go") {
			continue
		}

		currentContent, sha, err := handler.githubClient.GetFileContent(
			botgithub.GetFileContentArgs{
				Filename: *file.Filename,
				Owner:    handler.owner,
				Ref:      *pr.Head.Ref,
				Repo:     handler.repo,
			},
		)
		if err != nil {
			return fmt.Errorf("getting file content: %w", err)
		}

		updatedContent, err := handler.aiClient.ModifyCode(currentContent, changeRequest)
		if err != nil {
			return fmt.Errorf("AI modification failed: %w", err)
		}

		message := fmt.Sprintf("Update code based on feedback: %s", truncate(changeRequest, 50))

		if err := handler.githubClient.UpdateFile(
			botgithub.UpdateFileArgs{
				Branch:   *pr.Head.Ref,
				Content:  updatedContent,
				Filename: *file.Filename,
				Message:  message,
				Owner:    handler.owner,
				Repo:     handler.repo,
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

func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}

	return s[:length] + "..."
}
