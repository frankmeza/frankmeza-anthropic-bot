package botai

import "github.com/anthropics/anthropic-sdk-go"

func CreateMessageParams(prompt string) anthropic.MessageNewParams {
	return anthropic.MessageNewParams{
		MaxTokens: 5000,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
		Model: anthropic.ModelClaude3_7Sonnet20250219,
	}
}
