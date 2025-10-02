package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"

	botAi "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot_ai"
	botBlog "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot_blog"
	botCode "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot_code"
	botGithub "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot_github"
	"github.com/google/go-github/v57/github"
)

func healthCheck(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("OK"))
}

func main() {
	githubToken := os.Getenv("GITHUB_TOKEN")
	aiAPIKey := os.Getenv("AI_API_KEY")
	owner := os.Getenv("GITHUB_OWNER")
	repoWebsite := os.Getenv("GITHUB_REPO_WEBSITE")
	repoBot := os.Getenv("GITHUB_REPO_BOT")
	webhookSecret := os.Getenv("GITHUB_WEBHOOK_SECRET")

	if githubToken == "" || aiAPIKey == "" || owner == "" || repoWebsite == "" || repoBot == "" {
		log.Fatal("Missing required environment variables")
	}

	// create vendor client instances
	githubClient := botGithub.NewClient(githubToken)
	aiClient := botAi.NewClient(aiAPIKey)

	// ie which module will be used/handled?
	// pkg/bot_blog
	// pkg/bot_code

	blogHandler := botBlog.NewHandler(
		botBlog.Handler{
			AiClient:      aiClient,
			GithubClient:  githubClient,
			Owner:         owner,
			Repo:          repoWebsite,
			WebhookSecret: webhookSecret,
		},
	)

	codeHandler := botCode.NewHandler(githubClient, aiClient, owner, repoBot, webhookSecret)
	router := newRouter(blogHandler, codeHandler, repoWebsite, repoBot, webhookSecret)

	http.HandleFunc("/webhook", router.HandleWebhook)
	http.HandleFunc("/health", healthCheck)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("AI Blog Bot starting on :%s", port)
	log.Printf("Monitoring repos: %s/%s (blog), %s/%s (code)", owner, repoWebsite, owner, repoBot)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// router handles routing webhooks to the appropriate handler
type router struct {
	blogHandler   *botBlog.Handler
	codeHandler   *botCode.Handler
	repoWebsite   string
	repoBot       string
	webhookSecret string // Add this
}

// todo: create fn sig intf
func newRouter(blogHandler *botBlog.Handler, codeHandler *botCode.Handler, repoWebsite, repoBot, webhookSecret string) *router {
	return &router{
		blogHandler:   blogHandler,
		codeHandler:   codeHandler,
		repoWebsite:   repoWebsite,
		repoBot:       repoBot,
		webhookSecret: webhookSecret,
	}
}

func (router *router) HandleWebhook(writer http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(writer, "error reading body", http.StatusBadRequest)
		return
	}

	// Parse to get repo name (no validation yet)
	event, err := github.ParseWebHook(github.WebHookType(request), body)
	if err != nil {
		log.Printf("Webhook parsing failed: %v", err)
		http.Error(writer, "parsing failed", http.StatusBadRequest)
		return
	}

	var repoName string

	switch eventType := event.(type) {
	case *github.IssuesEvent:
		repoName = *eventType.Repo.FullName
	case *github.PullRequestReviewCommentEvent:
		repoName = *eventType.Repo.FullName
	}

	log.Printf("Detected repo: %s", repoName)

	// Recreate the request body for the handler
	request.Body = io.NopCloser(bytes.NewBuffer(body))

	switch {
	case contains(repoName, router.repoWebsite):
		log.Printf("Routing to blog handler")
		router.blogHandler.HandleWebhook(writer, request)

	case contains(repoName, router.repoBot):
		log.Printf("Routing to code handler")
		router.codeHandler.HandleWebhook(writer, request)

	default:
		log.Printf("Unknown repository: %s", repoName)
		writer.WriteHeader(http.StatusOK)
	}
}

func contains(parentString, childString string) bool {
	doesParentExist := len(parentString) > 0
	doesChildExist := len(childString) > 0

	areStringsEqual := parentString == childString

	// length of parentString minus the length of childString
	ideallyThisIsZeroIndex := len(parentString) - len(childString)

	// this value can be understood as using
	// - the difference in length as the beginning index (to the end with the colon character :)
	// - to compare that with the childString as-is for equality
	hasCharacterAndPositionEquality := parentString[ideallyThisIsZeroIndex:] == childString

	return doesParentExist && doesChildExist && (areStringsEqual || hasCharacterAndPositionEquality)
}
