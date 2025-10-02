package botblog

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	sharedUtils "github.com/frankmeza/frankmeza-anthropic-bot/pkg/shared_utils"
)

// BlogPostRequest represents data needed to create a blog post
type BlogPostRequest struct {
	Draft  bool     `json:"draft"`
	Points []string `json:"points"`
	Tags   []string `json:"tags"`
	Title  string   `json:"title"`
	Topic  string   `json:"topic"`
}

// Post represents a blog post with frontmatter matching your format
type Post struct {
	Content   string   `yaml:"-"`
	CreatedAt string   `yaml:"created_at"`
	IsDraft   bool     `yaml:"is_draft"`
	Key       string   `yaml:"key"`
	Language  string   `yaml:"language"`
	Summary   string   `yaml:"summary"`
	Tags      []string `yaml:"tags"`
	Title     string   `yaml:"title"`
	Type      string   `yaml:"type"`
}

// NewPost creates a new blog post with default values
func NewPost(title, topic string, tags []string, isDraft bool) *Post {
	key := generateKey(title)

	return &Post{
		CreatedAt: time.Now().Format("2000-12-31"),
		IsDraft:   isDraft,
		Key:       key,
		Language:  "en",
		Summary:   fmt.Sprintf("A casual exploration of %s", topic),
		Tags:      tags,
		Title:     title,
		Type:      "post",
	}
}

// FilePath returns the correct file path based on draft status
func (p *Post) GetFilePath() string {
	filename := fmt.Sprintf("%s.md", p.Key)

	if p.IsDraft {
		return filepath.Join("pkg", "blog_markdown_content", "drafts", filename)
	}

	return filepath.Join("pkg", "blog_markdown_content", "posts", filename)
}

// ToMarkdown converts the post to markdown format with frontmatter
func (p *Post) GenerateMarkdown() string {
	var buf bytes.Buffer

	buf.WriteString("---\n")
	buf.WriteString(fmt.Sprintf("created_at: %s\n", p.CreatedAt))
	buf.WriteString(fmt.Sprintf("is_draft: %t\n", p.IsDraft))
	buf.WriteString(fmt.Sprintf("key: %s\n", p.Key))
	buf.WriteString(fmt.Sprintf("language: %s\n", p.Language))
	buf.WriteString(fmt.Sprintf("summary: %s\n", p.Summary))

	buf.WriteString("tags:\n")
	for _, tag := range p.Tags {
		buf.WriteString(fmt.Sprintf("  - %s\n", tag))
	}

	buf.WriteString(fmt.Sprintf("title: %s\n", p.Title))
	buf.WriteString(fmt.Sprintf("type: %s\n", p.Type))
	buf.WriteString("---\n\n")
	buf.WriteString(p.Content)

	return buf.String()
}

// UpdateDraftStatus changes the draft status and updates the key if needed
func (p *Post) UpdateDraftStatus(isDraft bool) {
	p.IsDraft = isDraft
}

// generateKey creates a URL-friendly key from the title
func generateKey(title string) string {
	key := strings.ToLower(title)
	key = strings.ReplaceAll(key, " ", "-")

	// Remove special characters, keep only alphanumeric and hyphens
	var result strings.Builder
	for _, rune := range key {
		if sharedUtils.IsRuneDashCharacter(rune) ||
			sharedUtils.IsRuneNumerical(rune) ||
			sharedUtils.IsRuneDashCharacter(rune) {
			result.WriteRune(rune)
		}
	}

	return result.String()
}

// ParseIssueForRequest extracts blog post request data from GitHub issue
func ParseIssueForRequest(title, body string) *BlogPostRequest {
	// Remove "Blog post:" prefix if present
	cleanTitle := strings.TrimSpace(strings.TrimPrefix(title, "Blog post:"))

	request := &BlogPostRequest{
		Draft: true,                     // start as draft
		Tags:  []string{"ai-generated"}, // default tag
		Title: cleanTitle,
		Topic: body,
	}

	// Extract any mentioned tags from body
	if strings.Contains(strings.ToLower(body), "tags:") {
		// Simple tag extraction - look for "tags: golang, htmx, web"
		lines := strings.Split(body, "\n")

		for _, line := range lines {
			if strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "tags:") {
				tagsPart := strings.TrimPrefix(strings.ToLower(line), "tags:")
				tags := strings.Split(tagsPart, ",")

				for i, tag := range tags {
					tags[i] = strings.TrimSpace(tag)
				}

				request.Tags = append(request.Tags, tags...)
				break
			}
		}
	}

	return request
}
