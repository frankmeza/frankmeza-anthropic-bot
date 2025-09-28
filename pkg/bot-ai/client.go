package botai

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Client handles all AI operations using Anthropic's Claude
type Client struct {
	anthropic *anthropic.Client
}

// NewClient creates a new AI client with the provided API key
func NewClient(apiKey string) *Client {
	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &Client{
		anthropic: &client,
	}
}

// GenerateBlogPost creates blog post content based on the request
func (c *Client) GenerateBlogPost(request *BlogPostRequest) (string, error) {
	prompt := buildBlogPostPrompt(request)

	message, err := c.anthropic.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_7Sonnet20250219,
		MaxTokens: 2500,
		// MaxTokens: anthropic.Int(2500),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})

	if err != nil {
		return "", fmt.Errorf("anthropic API error: %w", err)
	}

	// Extract text from response
	if len(message.Content) > 0 {
		// if textBlock, ok := message.Content[0].(*anthropic.TextBlock); ok {
		// 	return textBlock.Text, nil
		// }

		textBlock := message.Content[0]
		return textBlock.Text, nil
	}

	return "", fmt.Errorf("unexpected response format from Anthropic")
}

// ModifyBlogPost updates existing blog post content based on feedback
func (c *Client) ModifyBlogPost(currentContent, changeRequest string) (string, error) {
	prompt := buildModificationPrompt(currentContent, changeRequest)

	message, err := c.anthropic.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_7Sonnet20250219,
		MaxTokens: 3000,
		// MaxTokens: anthropic.Int(3000),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})

	if err != nil {
		return "", fmt.Errorf("anthropic API error: %w", err)
	}

	// Extract text from response
	if len(message.Content) > 0 {
		// if textBlock, ok := message.Content[0].(*anthropic.TextBlock); ok {
		// 	return textBlock.Text, nil
		// }

		textBlock := message.Content[0]
		return textBlock.Text, nil
	}

	return "", fmt.Errorf("unexpected response format from Anthropic")
}

// BlogPostRequest represents the data needed to generate a blog post
type BlogPostRequest struct {
	Title  string   `json:"title"`
	Topic  string   `json:"topic"`
	Points []string `json:"points"`
	Tags   []string `json:"tags"`
	Draft  bool     `json:"draft"`
}
