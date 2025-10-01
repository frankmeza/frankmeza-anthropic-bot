package botai

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
)

// CodeRequest represents a request to generate code
type CodeRequest struct {
	Title       string
	Description string
	FileType    string
	TargetPath  string
	Tags        []string
}

// GenerateCode creates Go code based on the request
func (c *Client) GenerateCode(request *CodeRequest) (string, error) {
	prompt := buildCodeGenerationPrompt(request)

	message, err := c.anthropic.Messages.New(
		context.Background(),
		anthropic.MessageNewParams{
			MaxTokens: 5000,
			Model:     anthropic.ModelClaude3_7Sonnet20250219,
			Messages: []anthropic.MessageParam{
				anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
			},
		})

	if err != nil {
		return "", fmt.Errorf("anthropic API error: %w", err)
	}

	if len(message.Content) > 0 {
		textBlock := message.Content[0]
		return textBlock.Text, nil
	}

	return "", fmt.Errorf("unexpected response format from Anthropic")
}

// ModifyCode updates existing code based on feedback
func (c *Client) ModifyCode(currentContent, changeRequest string) (string, error) {
	prompt := buildCodeModificationPrompt(currentContent, changeRequest)

	message, err := c.anthropic.Messages.New(
		context.Background(),
		anthropic.MessageNewParams{
			MaxTokens: 5000,
			Model:     anthropic.ModelClaude3_7Sonnet20250219,
			Messages: []anthropic.MessageParam{
				anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
			},
		})

	if err != nil {
		return "", fmt.Errorf("anthropic API error: %w", err)
	}

	if len(message.Content) > 0 {
		textBlock := message.Content[0]
		return textBlock.Text, nil
	}

	return "", fmt.Errorf("unexpected response format from Anthropic")
}

// buildCodeGenerationPrompt creates the prompt for generating new code
func buildCodeGenerationPrompt(request *CodeRequest) string {
	return fmt.Sprintf(`You are an expert Go developer writing code for the frankmeza-anthropic-bot project. Generate Go code based on this request.

**Request:** %s

**Description:**
%s

**Target file:** %s

**Style Guidelines:**
- Follow Go best practices and idiomatic patterns
- Use clear, descriptive variable and function names
- Add blank lines between logical sections for readability
- Group related variable declarations at the top of functions
- Use early returns with blank lines for clarity
- Include error handling with descriptive error messages
- Add helpful comments for complex logic
- Match the existing code style in the project (see the bot-ai, bot-blog, bot-github packages)

**Code Structure:**
- If creating a new package, include package declaration
- Add necessary imports
- Define clear types and interfaces
- Implement functions with proper error handling
- Keep functions focused and single-purpose

Generate complete, working Go code that can be added to the project. Include only the code - no markdown code fences or explanations.`,
		request.Title,
		request.Description,
		request.TargetPath)
}

// buildCodeModificationPrompt creates the prompt for modifying existing code
func buildCodeModificationPrompt(currentContent, changeRequest string) string {
	return fmt.Sprintf(`You are an expert Go developer modifying code for the frankmeza-anthropic-bot project.

**Current code:**
%s

**Requested change:** "%s"

**Modification Guidelines:**
- Maintain the existing code style and structure
- Follow Go best practices and idiomatic patterns
- Preserve blank lines between logical sections
- Keep error handling patterns consistent
- Ensure changes are minimal and focused
- Add comments if the change adds complexity
- Test that the code compiles and makes sense

Return the complete modified code file. Include only the code - no markdown code fences or explanations.`,
		currentContent,
		changeRequest)
}
