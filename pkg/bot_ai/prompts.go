package botai

import (
	"fmt"
	"strings"
)

// buildBlogPostPrompt creates the prompt for generating new blog posts
func buildBlogPostPrompt(request *BlogPostRequest) string {
	return fmt.Sprintf(`You are a technical blog writer with a casual, clear writing style. Write a blog post about %s.

Style Guidelines:
- Casual, conversational tone but still informative and clear
- Include practical code examples in Go where relevant
- Use CSS classes in markdown format like: {.text-lg .text-gray-600 .mb-8}
- Start most paragraphs with appropriate CSS styling classes
- Include concrete, working examples that illustrate your points
- Keep it engaging and developer-friendly
- Write as if you're sharing knowledge with a fellow developer

Topic: %s
Key points to cover: %s
Target tags: %s

Write a complete blog post (just the content, no frontmatter) that would fit well on a developer's personal website. Include practical examples and maintain a light but informative tone.`,
		request.Topic,
		request.Topic,
		strings.Join(request.Points, ", "),
		strings.Join(request.Tags, ", "))
}

// buildModificationPrompt creates the prompt for modifying existing blog posts
func buildModificationPrompt(currentContent, changeRequest string) string {
	return fmt.Sprintf(`You are helping edit a blog post. A reader has requested a specific change to the content.

Current blog post:
%s

Change requested: "%s"

Please modify the blog post to address this request. Maintain the same:
- Frontmatter structure (don't change the YAML at the top)
- CSS class formatting like {.text-lg .text-gray-600 .mb-8}
- Casual, clear writing style
- Developer-friendly tone

Return the complete updated blog post including the original frontmatter.`,
		currentContent,
		changeRequest,
	)
}

// buildSummaryPrompt creates a prompt for generating blog post summaries
func buildSummaryPrompt(title, content string) string {
	return fmt.Sprintf(`Create a brief, engaging summary for this blog post:

Title: %s
Content: %s

Write a 1-2 sentence summary that captures the main topic and value for readers. Keep it casual but informative.`,
		title,
		content,
	)
}

// buildTagsPrompt creates a prompt for suggesting relevant tags
func buildTagsPrompt(title, content string) string {
	return fmt.Sprintf(`Suggest 3-5 relevant tags for this blog post:

Title: %s
Content: %s

Return only the tags as a comma-separated list. Focus on technical topics, programming languages, frameworks, and concepts mentioned.`,
		title,
		content,
	)
}
