package botcode

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ChangeRequest represents a code change request from an issue
type ChangeRequest struct {
	Title       string
	Description string
	FileType    string // "go", "md", etc.
	TargetPath  string // where the file should go
	Tags        []string
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

	// Extract target path if specified
	if strings.Contains(strings.ToLower(body), "file:") || strings.Contains(strings.ToLower(body), "path:") {
		lines := strings.Split(body, "\n")

		for _, line := range lines {
			lowerLine := strings.ToLower(strings.TrimSpace(line))

			if strings.HasPrefix(lowerLine, "file:") || strings.HasPrefix(lowerLine, "path:") {
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
		Path:    path,
		Content: content,
		Message: message,
	}
}

// IsGoFile checks if the file is a Go source file
func (cf *CodeFile) IsGoFile() bool {
	return strings.HasSuffix(cf.Path, ".go")
}

// Directory returns the directory portion of the path
func (cf *CodeFile) Directory() string {
	return filepath.Dir(cf.Path)
}

// Filename returns just the filename
func (cf *CodeFile) Filename() string {
	return filepath.Base(cf.Path)
}

// DetermineTargetPath figures out where a code file should go based on the request
func DetermineTargetPath(request *ChangeRequest) string {
	if request.TargetPath != "" {
		return request.TargetPath
	}

	// Default paths based on request type
	title := strings.ToLower(request.Title)

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
	return filepath.Join("pkg", "bot-code", filename)
}

// generateFilename creates a filename from a title
func generateFilename(title string) string {
	filename := strings.ToLower(title)
	filename = strings.ReplaceAll(filename, " ", "_")

	// Remove special characters
	var result strings.Builder

	for _, r := range filename {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			result.WriteRune(r)
		}
	}

	return result.String() + ".go"
}

// GenerateCommitMessage creates a descriptive commit message
func GenerateCommitMessage(request *ChangeRequest, action string) string {
	return fmt.Sprintf("%s: %s", action, request.Title)
}
