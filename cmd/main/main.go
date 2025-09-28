// AI Blog Bot for frankmeza/frankmeza repo
// This is the complete bot code - save as main.go
package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

type BlogPostRequest struct {
	Title  string   `json:"title"`
	Topic  string   `json:"topic"`
	Points []string `json:"points"`
	Tags   []string `json:"tags"`
	Draft  bool     `json:"draft"`
}

type BlogBot struct {
	githubClient *github.Client
	repo         string
	owner        string
	aiAPIKey     string
}

// BlogPost template structure matching your frontmatter
type BlogPost struct {
	CreatedAt string   `yaml:"created_at"`
	IsDraft   bool     `yaml:"is_draft"`
	Key       string   `yaml:"key"`
	Language  string   `yaml:"language"`
	Summary   string   `yaml:"summary"`
	Tags      []string `yaml:"tags"`
	Title     string   `yaml:"title"`
	Type      string   `yaml:"type"`
	Content   string   `yaml:"-"`
}

func NewBlogBot(githubToken, aiAPIKey, owner, repo string) *BlogBot {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubToken})
	tc := oauth2.NewClient(ctx, ts)

	return &BlogBot{
		githubClient: github.NewClient(tc),
		repo:         repo,
		owner:        owner,
		aiAPIKey:     aiAPIKey,
	}
}

func (b *BlogBot) handleWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, []byte(os.Getenv("GITHUB_WEBHOOK_SECRET")))
	if err != nil {
		log.Printf("webhook validation failed: %v", err)
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		log.Printf("webhook parsing failed: %v", err)
		return
	}

	switch e := event.(type) {
	case *github.IssuesEvent:
		if *e.Action == "opened" {
			b.handleNewIssue(e.Issue)
		}
	case *github.IssueCommentEvent:
		if *e.Action == "created" {
			b.handleIssueComment(e.Issue, e.Comment)
		}
	case *github.PullRequestReviewCommentEvent:
		if *e.Action == "created" {
			b.handlePRComment(e.PullRequest, e.Comment)
		}
	}
}

func (b *BlogBot) handleNewIssue(issue *github.Issue) {
	// Simple parsing - look for blog post requests
	title := *issue.Title
	body := *issue.Body

	if strings.Contains(strings.ToLower(title), "blog post") {
		// React with thumbs up
		b.reactToIssue(issue, "👍")

		// Parse the request and generate blog post
		request := b.parseIssueForBlogPost(title, body)
		if err := b.createBlogPostPR(issue, request); err != nil {
			log.Printf("Error creating blog post PR: %v", err)
			b.commentOnIssue(issue, "Sorry, I ran into an error creating the blog post. Could you check the request format?")
		}
	}
}

func (b *BlogBot) parseIssueForBlogPost(title, body string) *BlogPostRequest {
	// Basic parsing logic - you'd want to make this more sophisticated
	request := &BlogPostRequest{
		Title: strings.TrimPrefix(title, "Blog post: "),
		Topic: body,
		Tags:  []string{"ai-generated"}, // default tag
		Draft: true,                     // start as draft
	}

	// Extract any mentioned tags from body
	if strings.Contains(body, "tags:") {
		// Simple tag extraction logic
	}

	return request
}

func (b *BlogBot) createBlogPostPR(issue *github.Issue, request *BlogPostRequest) error {
	// Generate the blog post content using AI
	blogPost, err := b.generateBlogPost(request)
	if err != nil {
		return fmt.Errorf("generating blog post: %w", err)
	}

	// Create branch
	branchName := fmt.Sprintf("ai-blog-post-%d", *issue.Number)
	if err := b.createBranch(branchName); err != nil {
		return fmt.Errorf("creating branch: %w", err)
	}

	// Create markdown file in the correct directory based on draft status
	var filename string
	if blogPost.IsDraft {
		filename = fmt.Sprintf("pkg/blog_markdown_content/drafts/%s.md", blogPost.Key)
	} else {
		filename = fmt.Sprintf("pkg/blog_markdown_content/posts/%s.md", blogPost.Key)
	}
	content := b.formatBlogPostMarkdown(blogPost)

	if err := b.createFile(branchName, filename, content); err != nil {
		return fmt.Errorf("creating file: %w", err)
	}

	// Create PR
	if err := b.createPullRequest(branchName, issue, blogPost); err != nil {
		return fmt.Errorf("creating PR: %w", err)
	}

	return nil
}

func (b *BlogBot) generateBlogPost(request *BlogPostRequest) (*BlogPost, error) {
	// Generate content using AI API
	content, err := b.generateContentWithAI(request)
	if err != nil {
		// Fallback to template if AI fails
		content = b.generateMarkdownContent(request)
	}

	key := strings.ToLower(strings.ReplaceAll(request.Title, " ", "-"))
	key = strings.ReplaceAll(key, "[^a-z0-9-]", "") // clean up key

	blogPost := &BlogPost{
		CreatedAt: time.Now().Format("2006-01-02"),
		IsDraft:   request.Draft,
		Key:       key,
		Language:  "en",
		Summary:   fmt.Sprintf("A casual exploration of %s", request.Topic),
		Tags:      request.Tags,
		Title:     request.Title,
		Type:      "post",
		Content:   content,
	}

	return blogPost, nil
}

func (b *BlogBot) generateContentWithAI(request *BlogPostRequest) (string, error) {
	// This is where you'd integrate with OpenAI, Anthropic, etc.
	// Example structure for OpenAI:

	prompt := fmt.Sprintf(`Write a casual, clear blog post about %s.

Style guidelines:
- Light, conversational tone but still informative
- Include practical code examples in Go where relevant
- Use CSS classes in markdown like: {.text-lg .text-gray-600 .mb-8}
- Start paragraphs with appropriate CSS styling
- Include concrete examples that illustrate your points
- Keep it engaging but not overly chatty

Topic: %s
Key points to cover: %s

Write a complete blog post that would fit well on a developer's personal website.`,
		request.Topic, request.Topic, strings.Join(request.Points, ", "))

	// TODO: Replace with actual AI API call
	// Example: openai.CreateChatCompletion() or anthropic equivalent

	// For now, return an error so it falls back to template
	fmt.Println(prompt)
	return "", fmt.Errorf("AI integration not yet implemented")
}

func (b *BlogBot) generateMarkdownContent(request *BlogPostRequest) string {
	// This would call your AI API to generate actual content
	// For now, a basic template that matches your style
	content := fmt.Sprintf(`{.text-lg .text-gray-600 .mb-8}
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

	return content
}

func (b *BlogBot) formatBlogPostMarkdown(post *BlogPost) string {
	var buf bytes.Buffer

	buf.WriteString("---\n")
	buf.WriteString(fmt.Sprintf("created_at: %s\n", post.CreatedAt))
	buf.WriteString(fmt.Sprintf("is_draft: %t\n", post.IsDraft))
	buf.WriteString(fmt.Sprintf("key: %s\n", post.Key))
	buf.WriteString(fmt.Sprintf("language: %s\n", post.Language))
	buf.WriteString(fmt.Sprintf("summary: %s\n", post.Summary))
	buf.WriteString("tags:\n")
	for _, tag := range post.Tags {
		buf.WriteString(fmt.Sprintf("  - %s\n", tag))
	}
	buf.WriteString(fmt.Sprintf("title: %s\n", post.Title))
	buf.WriteString(fmt.Sprintf("type: %s\n", post.Type))
	buf.WriteString("---\n\n")
	buf.WriteString(post.Content)

	return buf.String()
}

// Helper methods for GitHub operations
func (b *BlogBot) reactToIssue(issue *github.Issue, reaction string) {
	ctx := context.Background()
	b.githubClient.Reactions.CreateIssueReaction(ctx, b.owner, b.repo, *issue.Number, reaction)
}

func (b *BlogBot) commentOnIssue(issue *github.Issue, comment string) {
	ctx := context.Background()
	b.githubClient.Issues.CreateComment(ctx, b.owner, b.repo, *issue.Number, &github.IssueComment{
		Body: &comment,
	})
}

func (b *BlogBot) createBranch(branchName string) error {
	ctx := context.Background()

	// Get the main branch reference
	mainRef, _, err := b.githubClient.Git.GetRef(ctx, b.owner, b.repo, "refs/heads/main")
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

	_, _, err = b.githubClient.Git.CreateRef(ctx, b.owner, b.repo, newRef)
	if err != nil {
		return fmt.Errorf("creating branch: %w", err)
	}

	return nil
}

func (b *BlogBot) createFile(branch, filename, content string) error {
	ctx := context.Background()

	opts := &github.RepositoryContentFileOptions{
		Message: github.String("Add AI-generated blog post"),
		Content: []byte(content),
		Branch:  github.String(branch),
	}

	_, _, err := b.githubClient.Repositories.CreateFile(ctx, b.owner, b.repo, filename, opts)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}

	return nil
}

func (b *BlogBot) createPullRequest(branch string, issue *github.Issue, post *BlogPost) error {
	ctx := context.Background()

	title := fmt.Sprintf("Add blog post: %s", post.Title)
	body := fmt.Sprintf(`🤖 AI-generated blog post based on issue #%d

**Title:** %s
**Summary:** %s
**Tags:** %s

This blog post was automatically generated. Feel free to comment with any changes you'd like me to make!

Closes #%d`, *issue.Number, post.Title, post.Summary, strings.Join(post.Tags, ", "), *issue.Number)

	head := fmt.Sprintf("%s:%s", b.owner, branch)
	base := "main"

	newPR := &github.NewPullRequest{
		Title: &title,
		Head:  &head,
		Base:  &base,
		Body:  &body,
	}

	_, _, err := b.githubClient.PullRequests.Create(ctx, b.owner, b.repo, newPR)
	if err != nil {
		return fmt.Errorf("creating PR: %w", err)
	}

	return nil
}

func (b *BlogBot) handleIssueComment(issue *github.Issue, comment *github.IssueComment) {
	// Handle comments on the original issue
}

func (b *BlogBot) handlePRComment(pr *github.PullRequest, comment *github.PullRequestComment) {
	// Handle comments on PR - this is where the magic happens!
	commentBody := *comment.Body

	// React with thumbs up to acknowledge
	ctx := context.Background()
	b.githubClient.Reactions.CreatePullRequestCommentReaction(ctx, b.owner, b.repo, *comment.ID, "👍")

	// Parse the comment for actionable requests
	if b.isChangeRequest(commentBody) {
		// Check for draft status changes
		if b.isDraftStatusChange(commentBody) {
			if err := b.handleDraftStatusChange(pr, commentBody); err != nil {
				log.Printf("Error changing draft status: %v", err)
			}
			return
		}

		// Get the current blog post content
		files, _, err := b.githubClient.PullRequests.ListFiles(ctx, b.owner, b.repo, *pr.Number, nil)
		if err != nil {
			log.Printf("Error getting PR files: %v", err)
			return
		}

		for _, file := range files {
			if strings.HasSuffix(*file.Filename, ".md") &&
				(strings.Contains(*file.Filename, "pkg/blog_markdown_content/posts") ||
					strings.Contains(*file.Filename, "pkg/blog_markdown_content/drafts")) {
				// This is our blog post file
				if err := b.updateBlogPost(pr, file, commentBody); err != nil {
					log.Printf("Error updating blog post: %v", err)
					b.commentOnPR(pr, "Sorry, I had trouble making that change. Could you be more specific?")
				} else {
					// React with rocket to show completion
					b.githubClient.Reactions.CreatePullRequestCommentReaction(ctx, b.owner, b.repo, *comment.ID, "🚀")
				}
				break
			}
		}
	}
}

func (b *BlogBot) isChangeRequest(comment string) bool {
	// Simple heuristics to detect change requests
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

func (b *BlogBot) updateBlogPost(pr *github.PullRequest, file *github.CommitFile, changeRequest string) error {
	ctx := context.Background()

	// Get current file content
	fileContent, _, _, err := b.githubClient.Repositories.GetContents(ctx, b.owner, b.repo, *file.Filename, &github.RepositoryContentGetOptions{
		Ref: *pr.Head.Ref,
	})
	if err != nil {
		return fmt.Errorf("getting file content: %w", err)
	}

	currentContent, err := fileContent.GetContent()
	if err != nil {
		return fmt.Errorf("decoding content: %w", err)
	}

	// Use AI to modify the content based on the change request
	updatedContent, err := b.modifyContentWithAI(currentContent, changeRequest)
	if err != nil {
		return fmt.Errorf("AI modification failed: %w", err)
	}

	// Update the file
	opts := &github.RepositoryContentFileOptions{
		Message: github.String(fmt.Sprintf("Update blog post based on feedback: %s", truncate(changeRequest, 50))),
		Content: []byte(updatedContent),
		Branch:  pr.Head.Ref,
		SHA:     fileContent.SHA,
	}

	_, _, err = b.githubClient.Repositories.UpdateFile(ctx, b.owner, b.repo, *file.Filename, opts)
	if err != nil {
		return fmt.Errorf("updating file: %w", err)
	}

	return nil
}

func (b *BlogBot) modifyContentWithAI(currentContent, changeRequest string) (string, error) {
	// This is where you'd integrate with your chosen AI API
	// For now, a simple placeholder that would be replaced with actual AI calls

	prompt := fmt.Sprintf(`Please modify this blog post based on the following request: "%s"

Current blog post content:
%s

Please return the complete updated blog post with the same frontmatter structure and CSS classes.`, changeRequest, currentContent)

	fmt.Println(prompt)

	// TODO: Replace with actual AI API call
	// For now, just return the original content with a note
	return currentContent + "\n\n{.text-sm .text-gray-500}\n*Note: This would be updated based on: " + changeRequest + "*\n", nil
}

func (b *BlogBot) commentOnPR(pr *github.PullRequest, comment string) {
	ctx := context.Background()
	b.githubClient.Issues.CreateComment(ctx, b.owner, b.repo, *pr.Number, &github.IssueComment{
		Body: &comment,
	})
}

func (b *BlogBot) isDraftStatusChange(comment string) bool {
	lowerComment := strings.ToLower(comment)
	return strings.Contains(lowerComment, "publish") ||
		strings.Contains(lowerComment, "ready to publish") ||
		strings.Contains(lowerComment, "move to draft") ||
		strings.Contains(lowerComment, "make it a draft")
}

func (b *BlogBot) handleDraftStatusChange(pr *github.PullRequest, comment string) error {
	ctx := context.Background()
	lowerComment := strings.ToLower(comment)

	// Get current files in the PR
	files, _, err := b.githubClient.PullRequests.ListFiles(ctx, b.owner, b.repo, *pr.Number, nil)
	if err != nil {
		return fmt.Errorf("getting PR files: %w", err)
	}

	for _, file := range files {
		if strings.HasSuffix(*file.Filename, ".md") &&
			(strings.Contains(*file.Filename, "pkg/blog_markdown_content/posts") ||
				strings.Contains(*file.Filename, "pkg/blog_markdown_content/drafts")) {

			// Determine new status and path
			var newFilename string
			var shouldPublish bool

			if strings.Contains(lowerComment, "publish") || strings.Contains(lowerComment, "ready to publish") {
				shouldPublish = true
				// Move from drafts to posts
				baseName := strings.TrimSuffix(filepath.Base(*file.Filename), ".md")
				newFilename = fmt.Sprintf("pkg/blog_markdown_content/posts/%s.md", baseName)
			} else {
				shouldPublish = false
				// Move from posts to drafts
				baseName := strings.TrimSuffix(filepath.Base(*file.Filename), ".md")
				newFilename = fmt.Sprintf("pkg/blog_markdown_content/drafts/%s.md", baseName)
			}

			// Get current content
			fileContent, _, _, err := b.githubClient.Repositories.GetContents(ctx, b.owner, b.repo, *file.Filename, &github.RepositoryContentGetOptions{
				Ref: *pr.Head.Ref,
			})
			if err != nil {
				return fmt.Errorf("getting file content: %w", err)
			}

			currentContent, err := fileContent.GetContent()
			if err != nil {
				return fmt.Errorf("decoding content: %w", err)
			}

			// Update the frontmatter
			updatedContent := b.updateDraftStatus(currentContent, !shouldPublish)

			// Create new file
			opts := &github.RepositoryContentFileOptions{
				Message: github.String(fmt.Sprintf("Move blog post to %s", map[bool]string{true: "published", false: "draft"}[shouldPublish])),
				Content: []byte(updatedContent),
				Branch:  pr.Head.Ref,
			}

			_, _, err = b.githubClient.Repositories.CreateFile(ctx, b.owner, b.repo, newFilename, opts)
			if err != nil {
				return fmt.Errorf("creating new file: %w", err)
			}

			// Delete old file
			deleteOpts := &github.RepositoryContentFileOptions{
				Message: github.String("Remove old blog post file"),
				Branch:  pr.Head.Ref,
				SHA:     fileContent.SHA,
			}

			_, _, err = b.githubClient.Repositories.DeleteFile(ctx, b.owner, b.repo, *file.Filename, deleteOpts)
			if err != nil {
				return fmt.Errorf("deleting old file: %w", err)
			}

			// Comment on the success
			statusMsg := map[bool]string{true: "published", false: "moved to drafts"}[shouldPublish]
			b.commentOnPR(pr, fmt.Sprintf("✅ Blog post %s!", statusMsg))

			break
		}
	}

	return nil
}

func (b *BlogBot) updateDraftStatus(content string, isDraft bool) string {
	// Update the is_draft field in the frontmatter
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		if strings.HasPrefix(line, "is_draft:") {
			lines[i] = fmt.Sprintf("is_draft: %t", isDraft)
			break
		}
	}

	return strings.Join(lines, " ")
}

func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}

func main() {
	githubToken := os.Getenv("GITHUB_TOKEN")
	aiAPIKey := os.Getenv("AI_API_KEY")
	owner := os.Getenv("GITHUB_OWNER")
	repo := os.Getenv("GITHUB_REPO")

	bot := NewBlogBot(githubToken, aiAPIKey, owner, repo)

	http.HandleFunc("/webhook", bot.handleWebhook)

	log.Println("AI Blog Bot starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
