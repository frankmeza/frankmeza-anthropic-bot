package shared

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

func TruncateText(textString string, limit int) string {
	if len(textString) <= limit {
		return textString
	}

	return textString[:limit] + "..."
}

func IsRuneAlphabetical(r rune) bool {
	return r >= 'a' && r <= 'z'
}

func IsRuneNumerical(r rune) bool {
	return r >= '0' && r <= '9'
}

func IsRuneDashCharacter(r rune) bool {
	return r == '-'
}
