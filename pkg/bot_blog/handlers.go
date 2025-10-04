package botblog

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	botAi "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot_ai"
	botGithub "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot_github"
	sharedUtils "github.com/frankmeza/frankmeza-anthropic-bot/pkg/shared_utils"
	"github.com/google/go-github/v57/github"
)

// Handler manages webhook events and blog operations
type Handler struct {
	AiClient      *botAi.Client
	GithubClient  *botGithub.Client
	Owner         string
	Repo          string
	WebhookSecret string
}

// NewHandler creates a new blog handler
func NewHandler(args Handler) *Handler {
	return &Handler{
		AiClient:      args.AiClient,
		GithubClient:  args.GithubClient,
		Owner:         args.Owner,
		Repo:          args.Repo,
		WebhookSecret: args.WebhookSecret,
	}
}

// HandleWebhook processes GitHub webhook events
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

	event, err := github.ParseWebHook(github.WebHookType(request), payload)
	if err != nil {
		log.Printf("webhook parsing failed: %v", err)
		http.Error(writer, "parsing failed", http.StatusBadRequest)
		return
	}

	switch e := event.(type) {
	case *github.IssuesEvent:
		if *e.Action == "opened" {
			handler.handleNewIssue(e.Issue)
		}
	case *github.IssueCommentEvent:
		if *e.Action == "created" {
			handler.handleIssueComment(e.Issue, e.Comment)
		}
	case *github.PullRequestReviewCommentEvent:
		if *e.Action == "created" {
			handler.handlePRComment(e.PullRequest, e.Comment)
		}
	}

	writer.WriteHeader(http.StatusOK)
}

// handleNewIssue processes new GitHub issues
func (handler *Handler) handleNewIssue(issue *github.Issue) {
	title := *issue.Title
	body := *issue.Body

	// Check if this is a blog post request
	if !strings.Contains(strings.ToLower(title), "blog post") {
		return
	}

	// React with thumbs up to acknowledge
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

	// Parse the request and generate blog post
	request := ParseIssueForRequest(title, body)
	if err := handler.createBlogPostPR(issue, request); err != nil {
		log.Printf("Error creating blog post PR: %v", err)
		handler.GithubClient.CommentOnIssue(
			botGithub.CommentOnIssueArgs{
				Comment:     "Sorry, I ran into an error creating the blog post. Could you check the request format?",
				IssueNumber: *issue.Number,
				Owner:       handler.Owner,
				Repo:        handler.Repo,
			},
		)
	}
}

// createBlogPostPR generates a blog post and creates a PR
func (handler *Handler) createBlogPostPR(issue *github.Issue, request *BlogPostRequest) error {
	// Generate the blog post content using AI
	content, err := handler.AiClient.GenerateBlogPost(
		&botAi.BlogPostRequest{
			Title:  request.Title,
			Topic:  request.Topic,
			Points: request.Points,
			Tags:   request.Tags,
			Draft:  request.Draft,
		},
	)

	if err != nil {
		log.Printf("AI generation failed, using template: %v", err)
		content = handler.generateTemplateContent(request)
	}

	// instantiate blog post struct, incomplete
	post := NewPost(
		request.Title,
		request.Topic,
		request.Tags,
		request.Draft,
	)

	// post content is assigned here
	post.Content = content

	// Create branch
	branchName := fmt.Sprintf("ai-assisted-post-%d", *issue.Number)

	if err := handler.GithubClient.CreateBranch(
		botGithub.CreateBranchArgs{
			BranchName: branchName,
			Owner:      handler.Owner,
			Repo:       handler.Repo,
		},
	); err != nil {
		return fmt.Errorf("creating branch: %w", err)
	}

	// Create markdown file
	filename := post.GetFilePath()
	markdown := post.GenerateMarkdown()
	message := "Add AI-generated blog post"

	if err := handler.GithubClient.CreateFile(
		botGithub.CreateFileArgs{
			Branch:   branchName,
			Content:  markdown,
			Filename: filename,
			Message:  message,
			Owner:    handler.Owner,
			Repo:     handler.Repo,
		},
	); err != nil {
		return fmt.Errorf("creating file: %w", err)
	}

	// Create PR
	title := fmt.Sprintf("Add blog post: %s", post.Title)
	body := handler.generatePRBody(issue, post)
	head := fmt.Sprintf("%s:%s", handler.Owner, branchName)

	_, err = handler.GithubClient.CreatePullRequest(
		botGithub.CreatePullRequestArgs{
			Body:  body,
			Base:  "main",
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

// handlePRComment processes comments on pull requests
func (handler *Handler) handlePRComment(
	pullRequest *github.PullRequest,
	comment *github.PullRequestComment,
) {
	// React with thumbs up to acknowledge
	if err := handler.GithubClient.ReactToPRComment(
		botGithub.ReactToPRCommentArgs{
			CommentID: *comment.ID,
			Owner:     handler.Owner,
			Reaction:  "+1",
			Repo:      handler.Repo,
		},
	); err != nil {
		log.Printf("Error reacting to PR comment: %v", err)
	}

	commentBody := *comment.Body

	// Check for draft status changes
	if handler.hasDraftStatusChange(commentBody) {
		if err := handler.handleDraftStatusChange(pullRequest, commentBody); err != nil {
			log.Printf("Error changing draft status: %v", err)
		}

		return
	}

	// Handle content changes
	if handler.isChangeRequest(commentBody) {
		if err := handler.handleContentChange(pullRequest, commentBody); err != nil {
			log.Printf("Error updating content: %v", err)

			handler.GithubClient.CommentOnPR(
				botGithub.CommentOnPRArgs{
					Comment:  "Sorry, I had trouble making that change. Could you be more specific?",
					Owner:    handler.Owner,
					PrNumber: *pullRequest.Number,
					Repo:     handler.Repo,
				},
			)
		} else {
			// React with rocket to show completion
			handler.GithubClient.ReactToPRComment(
				botGithub.ReactToPRCommentArgs{
					CommentID: *comment.ID,
					Owner:     handler.Owner,
					Reaction:  "rocket",
					Repo:      handler.Repo,
				},
			)
		}
	}
}

// handleContentChange modifies blog post content based on feedback
func (handler *Handler) handleContentChange(
	pullRequest *github.PullRequest,
	changeRequest string,
) error {
	// Get files changed in this PR
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

	// Find the blog post file
	for _, file := range files {
		isMarkdownFile := strings.HasSuffix(*file.Filename, ".md")
		isFileInPostsDir := strings.Contains(*file.Filename, "pkg/blog_markdown_content/posts")
		isFileInDraftsDir := strings.Contains(*file.Filename, "pkg/blog_markdown_content/drafts")

		if isMarkdownFile && (isFileInPostsDir || isFileInDraftsDir) {
			// Get current content
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

			// Use AI to modify the content
			updatedContent, err := handler.AiClient.ModifyBlogPost(
				currentContent,
				changeRequest,
			)

			if err != nil {
				return fmt.Errorf("AI modification failed: %w", err)
			}

			// Update the file
			message := fmt.Sprintf(
				"Update blog post based on feedback: %s",
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
	}

	return nil
}

// handleDraftStatusChange moves blog posts between drafts and posts directories
func (handler *Handler) handleDraftStatusChange(
	pullRequest *github.PullRequest,
	comment string,
) error {
	lowerComment := strings.ToLower(comment)

	shouldPublish := strings.Contains(lowerComment, "publish") ||
		strings.Contains(lowerComment, "ready to publish")

	// Get files in the PR
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
		isMarkdownFile := strings.HasSuffix(*file.Filename, ".md")
		isFileInPostsDir := strings.Contains(*file.Filename, "pkg/blog_markdown_content/posts")
		isFileInDraftsDir := strings.Contains(*file.Filename, "pkg/blog_markdown_content/drafts")

		if isMarkdownFile && (isFileInPostsDir || isFileInDraftsDir) {
			// Get current content
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

			// Update draft status in content
			updatedContent := handler.updateDraftStatus(currentContent, !shouldPublish)

			// Determine new file path
			baseName := strings.TrimSuffix(filepath.Base(*file.Filename), ".md")

			var newFilename string
			if shouldPublish {
				newFilename = filepath.Join("pkg", "blog_markdown_content", "posts", baseName+".md")
			} else {
				newFilename = filepath.Join("pkg", "blog_markdown_content", "drafts", baseName+".md")
			}

			// Create new file in /posts
			message := fmt.Sprintf(
				"Move blog post to %s",
				map[bool]string{true: "published", false: "draft"}[shouldPublish],
			)

			if err := handler.GithubClient.CreateFile(
				botGithub.CreateFileArgs{
					Branch:   *pullRequest.Head.Ref,
					Content:  updatedContent,
					Filename: newFilename,
					Message:  message,
					Owner:    handler.Owner,
					Repo:     handler.Repo,
				},
			); err != nil {
				return fmt.Errorf("creating new file: %w", err)
			}

			// Delete old file in /drafts assumedly
			if err := handler.GithubClient.DeleteFile(
				botGithub.DeleteFileArgs{
					Owner:    handler.Owner,
					Repo:     handler.Repo,
					Branch:   *pullRequest.Head.Ref,
					Filename: *file.Filename,
					Message:  "Remove old blog post file",
					Sha:      sha,
				},
			); err != nil {
				return fmt.Errorf("deleting old file: %w", err)
			}

			// Comment on success
			statusMsg := map[bool]string{
				true:  "published",
				false: "moved to drafts",
			}[shouldPublish]

			handler.GithubClient.CommentOnPR(
				botGithub.CommentOnPRArgs{
					Comment:  fmt.Sprintf("✅ Blog post %s!", statusMsg),
					Owner:    handler.Owner,
					PrNumber: *pullRequest.Number,
					Repo:     handler.Repo,
				},
			)

			break
		}
	}

	return nil
}

// Helper methods

func (handler *Handler) handleIssueComment(
	issue *github.Issue,
	comment *github.IssueComment,
) {
	// Handle comments on the original issue if needed
	// For now, we mainly focus on PR comments
}

func (handler *Handler) isChangeRequest(comment string) bool {
	changeWords := []string{
		"can you", "could you", "please", "add", "remove", "change", "update",
		"make it", "make this", "more", "less", "fix", "improve", "rewrite",
	}

	lowerComment := strings.ToLower(comment)
	for _, word := range changeWords {
		if strings.Contains(lowerComment, word) {
			return true
		}
	}
	return false
}

func (handler *Handler) hasDraftStatusChange(comment string) bool {
	lowerComment := strings.ToLower(comment)
	return strings.Contains(lowerComment, "publish") ||
		strings.Contains(lowerComment, "ready to publish") ||
		strings.Contains(lowerComment, "move to draft") ||
		strings.Contains(lowerComment, "make it a draft")
}

func (handler *Handler) updateDraftStatus(content string, isDraft bool) string {
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		if strings.HasPrefix(line, "is_draft:") {
			lines[i] = fmt.Sprintf("is_draft: %t", isDraft)
			break
		}
	}

	return strings.Join(lines, "\n")
}

func (handler *Handler) generatePRBody(issue *github.Issue, post *Post) string {
	return fmt.Sprintf(`🤖 AI-generated blog post based on issue #%d

**Title:** %s
**Summary:** %s
**Tags:** %s

This blog post was automatically generated. Feel free to comment with any changes you'd like me to make!

Closes #%d`, *issue.Number, post.Title, post.Summary, strings.Join(post.Tags, ", "), *issue.Number)
}

func (handler *Handler) generateTemplateContent(request *BlogPostRequest) string {
	return fmt.Sprintf(`{.text-lg .text-gray-600 .mb-8}
Hey there! Let's dive into %s - it's one of those topics that's both fascinating and practical.

{.text-base .mb-6}
So what exactly are we talking about here? %s is something I've been exploring lately, and I thought it'd be fun to share some insights.

{.text-base .mb-6}
Here's a quick example to get us started:

~~~go
// Simple example illustrating the concept
func example() {
    fmt.Println("This is where we'd show some code!")
}
~~~

{.text-base .mb-6}
Pretty neat, right? The cool thing about this approach is how it keeps things simple while still being powerful.

{.text-base .mb-6}
What I find interesting is how this connects to broader patterns in development. It's not just about the technical details - it's about building something that actually works well in practice.

{.text-base .mb-6}
Hope this gives you some ideas to play around with! As always, feel free to reach out if you want to chat more about this stuff.
`, request.Topic, request.Topic)
}
