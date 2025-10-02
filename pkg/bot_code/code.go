package botcode

import (
	"fmt"
	"path/filepath"
	"strings"

	sharedUtils "github.com/frankmeza/frankmeza-anthropic-bot/pkg/shared_utils"
)

// ChangeRequest represents a code change request from an issue
type ChangeRequest struct {
	Description string
	FileType    string // "go", "md", etc.
	Tags        []string
	TargetPath  string // where the file should go
	Title       string
}

// ParseIssueForCodeRequest extracts code change request data from GitHub issue
func ParseIssueForCodeRequest(title, body string) *ChangeRequest {
	cleanTitle := strings.TrimSpace(strings.TrimPrefix(title, "Code:"))
	cleanTitle = strings.TrimSpace(strings.TrimPrefix(cleanTitle, "code:"))

	request := &ChangeRequest{
		Title:       cleanTitle,
		Description: body,
		FileType:    "go",
		Tags:        []string{"ai-generated"},
	}

	doesBodyContainFileKeyword := strings.Contains(strings.ToLower(body), "file:")
	doesBodyContainPathKeyword := strings.Contains(strings.ToLower(body), "path:")

	// Extract target path if specified
	if doesBodyContainFileKeyword || doesBodyContainPathKeyword {
		lines := strings.SplitSeq(body, "\n")

		for line := range lines {
			lowerLine := strings.ToLower(strings.TrimSpace(line))

			doesFileHavePrefixFileKeyword := strings.HasPrefix(lowerLine, "file:")
			doesFileHavePrefixPathKeyword := strings.HasPrefix(lowerLine, "path:")

			if doesFileHavePrefixFileKeyword || doesFileHavePrefixPathKeyword {
				parts := strings.SplitN(line, ":", 2)

				if len(parts) == 2 {
					request.TargetPath = strings.TrimSpace(parts[1])
				}
			}
		}
	}

	return request
}

// CodeFile represents a Go code file to be created or modified
type CodeFile struct {
	Path    string
	Content string
	Message string // commit message
}

// NewCodeFile creates a new code file structure
func NewCodeFile(path, content, message string) *CodeFile {
	return &CodeFile{
		Content: content,
		Message: message,
		Path:    path,
	}
}

// IsGoFile checks if the file is a Go source file
func (codeFile *CodeFile) IsGoFile() bool {
	return strings.HasSuffix(codeFile.Path, ".go")
}

// Directory returns the directory portion of the path
func (codeFile *CodeFile) GetDirectory() string {
	return filepath.Dir(codeFile.Path)
}

// GetFilePath returns just the filename
func (codeFile *CodeFile) GetFilePath() string {
	return filepath.Base(codeFile.Path)
}

// DetermineTargetPath figures out where a code file should go based on the request
// todo - oh this needs help
func DetermineTargetPath(request *ChangeRequest) string {
	if request.TargetPath != "" {
		return request.TargetPath
	}

	// Default paths based on request type
	title := strings.ToLower(request.Title)

	// todo this needs to be addressed, it's not amazing
	if strings.Contains(title, "handler") {
		return "pkg/bot-code/handlers.go"
	}

	if strings.Contains(title, "client") {
		return "pkg/bot-code/client.go"
	}

	if strings.Contains(title, "test") {
		return "pkg/bot-code/code_test.go"
	}

	// Default to a new file based on title
	filename := generateFilename(request.Title)

	// todo wut is this
	return filepath.Join("pkg", "bot-code", filename)
}

// generateFilename creates a filename from a title
func generateFilename(title string) string {
	filename := strings.ToLower(title)
	filename = strings.ReplaceAll(filename, " ", "_")

	// Remove special characters
	var result strings.Builder

	for _, rune := range filename {
		if sharedUtils.IsRuneAlphabetical(rune) ||
			sharedUtils.IsRuneNumerical(rune) ||
			sharedUtils.IsRuneDashCharacter(rune) {
			result.WriteRune(rune)
		}
	}

	return result.String() + ".go"
}

// GenerateCommitMessage creates a descriptive commit message
func GenerateCommitMessage(request *ChangeRequest, action string) string {
	return fmt.Sprintf("%s: %s", action, request.Title)
}
