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
	context   context.Context
}

// BlogPostRequest represents the data needed to generate a blog post
type BlogPostRequest struct {
	Draft  bool     `json:"draft"`
	Points []string `json:"points"`
	Tags   []string `json:"tags"`
	Title  string   `json:"title"`
	Topic  string   `json:"topic"`
}

// NewClient creates a new AI client with the provided API key
func NewClient(apiKey string) *Client {
	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &Client{
		anthropic: &client,
		context:   context.Background(),
	}
}

// GenerateBlogPost creates blog post content based on the request
func (client *Client) GenerateBlogPost(request *BlogPostRequest) (string, error) {
	prompt := buildBlogPostPrompt(request)

	message, err := client.anthropic.Messages.New(
		context.Background(),
		anthropic.MessageNewParams{
			MaxTokens: 5000,
			Model:     anthropic.ModelClaude3_7Sonnet20250219,
			Messages: []anthropic.MessageParam{
				anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("anthropic API error: %w", err)
	}

	// Extract text from response
	if len(message.Content) > 0 {
		textBlock := message.Content[0]
		return textBlock.Text, nil
	}

	return "", fmt.Errorf("unexpected response format from Anthropic")
}

// ModifyBlogPost updates existing blog post content based on feedback
func (client *Client) ModifyBlogPost(
	currentContent string,
	changeRequest string,
) (string, error) {
	prompt := buildModificationPrompt(currentContent, changeRequest)

	message, err := client.anthropic.Messages.New(
		context.Background(),
		anthropic.MessageNewParams{
			MaxTokens: 5000,
			Model:     anthropic.ModelClaude3_7Sonnet20250219,
			Messages: []anthropic.MessageParam{
				anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("anthropic API error: %w", err)
	}

	// Extract text from response
	if len(message.Content) > 0 {
		textBlock := message.Content[0]
		return textBlock.Text, nil
	}

	return "", fmt.Errorf("unexpected response format from Anthropic")
}
