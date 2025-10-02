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