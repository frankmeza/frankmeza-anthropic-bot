// Package constants provides centralized string constants for the frankmeza-anthropic-bot project
package constants

// Common messages
const (
	InfoPrefix     = "[INFO] "
	WarningPrefix  = "[WARNING] "
	ErrorPrefix    = "[ERROR] "
	SuccessMessage = "Operation completed successfully"
	FailureMessage = "Operation failed"
)

// Error messages
const (
	ErrInvalidInput     = "invalid input provided"
	ErrNotFound         = "resource not found"
	ErrUnauthorized     = "unauthorized access"
	ErrInternalServer   = "internal server error"
	ErrTimeout          = "operation timed out"
	ErrInvalidOperation = "invalid operation"
	ErrEmptyResponse    = "empty response received"
)

// HTTP related
const (
	ContentTypeJSON           = "application/json"
	ContentTypeText           = "text/plain"
	HeaderContentType         = "Content-Type"
	HeaderAuthorization       = "Authorization"
	HeaderUserAgent           = "User-Agent"
	DefaultUserAgent          = "frankmeza-anthropic-bot"
	AuthBearer                = "Bearer "
	RequestMethodGET          = "GET"
	RequestMethodPOST         = "POST"
	RequestMethodPUT          = "PUT"
	RequestMethodDELETE       = "DELETE"
	StatusOK                  = "200 OK"
	StatusCreated             = "201 Created"
	StatusBadRequest          = "400 Bad Request"
	StatusUnauthorized        = "401 Unauthorized"
	StatusNotFound            = "404 Not Found"
	StatusInternalServerError = "500 Internal Server Error"
)

// API related
const (
	AnthropicAPIBaseURL     = "https://api.anthropic.com"
	AnthropicAPIVersion     = "2023-06-01"
	GithubAPIBaseURL        = "https://api.github.com"
	DefaultMaxTokens        = 4096
	DefaultTemperature      = 0.7
	DefaultModel            = "claude-2.1"
	AnthropicHeaderVersion  = "anthropic-version"
	AnthropicHeaderBetaPath = "anthropic-beta"
)

// Configuration keys
const (
	ConfigAnthropicAPIKey  = "ANTHROPIC_API_KEY"
	ConfigGithubToken      = "GITHUB_TOKEN"
	ConfigLogLevel         = "LOG_LEVEL"
	ConfigPort             = "PORT"
	ConfigHost             = "HOST"
	ConfigEnvironment      = "ENVIRONMENT"
	ConfigDevEnvironment   = "development"
	ConfigProdEnvironment  = "production"
	ConfigTestEnvironment  = "test"
	ConfigDefaultPort      = "8080"
	ConfigDefaultHost      = "0.0.0.0"
	ConfigDefaultLogLevel  = "info"
)

// File paths and names
const (
	DefaultConfigPath    = "./config.yaml"
	LogFilePath          = "./logs/app.log"
	TempDirectoryPath    = "./temp"
	DefaultOutputPath    = "./output"
	DefaultTemplatesPath = "./templates"
)

// Time formats
const (
	TimeFormatISO8601        = "2006-01-02T15:04:05Z07:00"
	TimeFormatSimple         = "2006-01-02 15:04:05"
	TimeFormatDateOnly       = "2006-01-02"
	TimeFormatTimeOnly       = "15:04:05"
	TimeFormatCompactDate    = "20060102"
	TimeFormatCompactTime    = "150405"
	TimeFormatCompactFull    = "20060102150405"
	TimeFormatHumanReadable  = "January 2, 2006 at 3:04 PM"
)

// Chat and prompt related
const (
	DefaultSystemPrompt     = "You are a helpful AI assistant."
	DefaultUserPromptPrefix = "User: "
	DefaultBotPromptPrefix  = "Assistant: "
	PromptSeparator         = "\n\n"
	MaxAttempts             = 3
	DefaultContextWindow    = 8192
)

// Database related
const (
	DBConnectionStringFormat = "%s:%s@tcp(%s:%s)/%s?parseTime=true"
	DBDefaultHost            = "localhost"
	DBDefaultPort            = "3306"
	DBDefaultUser            = "root"
	DBDefaultDatabase        = "anthropic_bot"
)

// Command line flags
const (
	FlagVerbose      = "verbose"
	FlagConfig       = "config"
	FlagVersion      = "version"
	FlagHelp         = "help"
	FlagOutputFormat = "output"
	FlagQuiet        = "quiet"
)

// Miscellaneous
const (
	DefaultTimeout         = 30
	MaxRetries             = 3
	RetryDelay             = 2
	DefaultBatchSize       = 100
	EmptyString            = ""
	SpaceString            = " "
	CommaString            = ","
	NewLineString          = "\n"
	TabString              = "\t"
	DefaultSeparator       = ","
	DefaultLanguage        = "en"
	DefaultEncoding        = "utf-8"
	ProjectName            = "frankmeza-anthropic-bot"
	LogPrefix              = "[ANTHROPIC-BOT] "
	VersionFormat          = "v%s.%s.%s"
	DefaultDebounceTime    = 500
	MaxConcurrentRequests  = 5
	DefaultCacheTTL        = 3600
)

// BotAI specific constants
const (
	AIDefaultPrompt           = "Please provide a detailed and helpful response."
	AISystemRolePrompt        = "You are an AI assistant specialized in providing accurate and helpful information."
	AIErrorGeneratingResponse = "Error generating AI response"
	AIResponsePrefix          = "AI Response: "
	AITokenLimitExceeded      = "Token limit exceeded, truncating response"
	AIGeneratingResponse      = "Generating AI response..."
	AIModelUnavailable        = "The specified AI model is currently unavailable"
	AIInvalidPrompt           = "The provided prompt is invalid or too short"
)

// BotBlog specific constants
const (
	BlogPostTemplate          = "# {title}\n\n{content}\n\n*Posted on {date}*"
	BlogGeneratePrompt        = "Write a blog post about the following topic: "
	BlogDefaultTitle          = "Untitled Blog Post"
	BlogDefaultCategory       = "General"
	BlogErrorGeneratingPost   = "Error generating blog post"
	BlogPostSaved             = "Blog post saved successfully"
	BlogInvalidTitle          = "Invalid blog post title"
	BlogInvalidContent        = "Invalid blog post content"
	BlogPostNotFound          = "Blog post not found"
	BlogCategoryNotFound      = "Blog category not found"
	BlogGeneratingPost        = "Generating blog post..."
	BlogFormattingPost        = "Formatting blog post..."
	BlogPostGenerationPrompt  = "Create an informative and engaging blog post about: "
	BlogPostFormatInstruction = "Format the post with markdown, including headers, lists, and emphasis where appropriate."
)

// BotCode specific constants
const (
	CodeReviewPrompt           = "Review the following code and provide feedback: "
	CodeGeneratePrompt         = "Generate code for the following requirements: "
	CodeExplainPrompt          = "Explain the following code in detail: "
	CodeOptimizePrompt         = "Optimize the following code: "
	CodeDefaultLanguage        = "go"
	CodeErrorGeneratingCode    = "Error generating code"
	CodeErrorReviewingCode     = "Error reviewing code"
	CodeErrorExplainingCode    = "Error explaining code"
	CodeErrorOptimizingCode    = "Error optimizing code"
	CodeGeneratingResponse     = "Generating code response..."
	CodeReviewingCode          = "Reviewing code..."
	CodeExplainingCode         = "Explaining code..."
	CodeOptimizingCode         = "Optimizing code..."
	CodeInvalidInput           = "Invalid code input"
	CodeBlockStart             = "```"
	CodeBlockEnd               = "```"
	CodeLanguagePrefix         = "```"
	CodeReviewSummaryHeader    = "# Code Review Summary"
	CodeOptimizationHeader     = "# Optimization Suggestions"
	CodeExplanationHeader      = "# Code Explanation"
	CodeGeneratedHeader        = "# Generated Code"
	CodeReviewTemplate         = "## Issues\n\n{issues}\n\n## Suggestions\n\n{suggestions}"
	CodeDocumentationTemplate  = "/**\n * {description}\n * @param {params}\n * @return {returns}\n */"
)

// BotGitHub specific constants
const (
	GitHubAPIReposURL             = "/repos/"
	GitHubAPIIssuesURL            = "/issues"
	GitHubAPIPullsURL             = "/pulls"
	GitHubAPIContentsURL          = "/contents/"
	GitHubDefaultBranch           = "main"
	GitHubErrorFetchingRepo       = "Error fetching repository information"
	GitHubErrorFetchingIssues     = "Error fetching issues"
	GitHubErrorFetchingPulls      = "Error fetching pull requests"
	GitHubErrorFetchingContents   = "Error fetching repository contents"
	GitHubErrorCreatingIssue      = "Error creating issue"
	GitHubErrorCreatingPR         = "Error creating pull request"
	GitHubFetchingRepoInfo        = "Fetching repository information..."
	GitHubFetchingIssues          = "Fetching repository issues..."
	GitHubFetchingPulls           = "Fetching repository pull requests..."
	GitHubFetchingContents        = "Fetching repository contents..."
	GitHubCreatingIssue           = "Creating new issue..."
	GitHubCreatingPR              = "Creating new pull request..."
	GitHubInvalidRepo             = "Invalid repository name or owner"
	GitHubInvalidIssueTitle       = "Invalid issue title"
	GitHubInvalidPRTitle          = "Invalid pull request title"
	GitHubIssueCreated            = "Issue created successfully"
	GitHubPRCreated               = "Pull request created successfully"
	GitHubDefaultIssueBody        = "This issue was created by the frankmeza-anthropic-bot."
	GitHubDefaultPRBody           = "This pull request was created by the frankmeza-anthropic-bot."
	GitHubDefaultCommitMessage    = "Update from frankmeza-anthropic-bot"
	GitHubDefaultPRTitle          = "Automated PR from frankmeza-anthropic-bot"
	GitHubDefaultIssueTitle       = "Automated Issue from frankmeza-anthropic-bot"
	GitHubRepoNotFound            = "Repository not found"
	GitHubUnauthorized            = "Unauthorized GitHub access"
	GitHubRateLimitExceeded       = "GitHub API rate limit exceeded"
	GitHubFileNotFound            = "File not found in repository"
)