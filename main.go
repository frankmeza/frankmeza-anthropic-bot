package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	botAi "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot_ai"
	botBlog "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot_blog"
	botCode "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot_code"
	botGithub "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot_github"
	"github.com/google/go-github/v57/github"
	"github.com/joho/godotenv"
)

func healthCheck(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("OK"))
}

func main() {
	envFile, _ := godotenv.Read(".env")

	for key, value := range envFile {
		if value == "" {
			fmt.Sprint("Missing required environment variable ", key)
		}
	}

	// create vendor client instances
	githubClient := botGithub.NewClient(envFile["GITHUB_TOKEN"])
	aiClient := botAi.NewClient(envFile["AI_API_KEY"])

	blogHandler := botBlog.NewHandler(
		botBlog.Handler{
			AiClient:      aiClient,
			GithubClient:  githubClient,
			Owner:         envFile["GITHUB_OWNER"],
			Repo:          envFile["GITHUB_REPO_WEBSITE"],
			WebhookSecret: envFile["GITHUB_WEBHOOK_SECRET"],
		},
	)

	codeHandler := botCode.NewHandler(
		botCode.Handler{
			AiClient:      aiClient,
			GithubClient:  githubClient,
			Owner:         envFile["GITHUB_OWNER"],
			Repo:          envFile["GITHUB_REPO_WEBSITE"],
			WebhookSecret: envFile["GITHUB_WEBHOOK_SECRET"],
		},
	)

	router := newRouter(
		router{
			blogHandler:   blogHandler,
			codeHandler:   codeHandler,
			repoWebsite:   envFile["GITHUB_REPO_WEBSITE"],
			repoBot:       envFile["GITHUB_REPO_BOT"],
			webhookSecret: envFile["GITHUB_WEBHOOK_SECRET"],
		},
	)

	http.HandleFunc("/webhook", router.HandleWebhook)
	http.HandleFunc("/health", healthCheck)

	port := "8080"

	log.Printf("AI Blog Bot starting on :%s", port)

	log.Printf(
		"Monitoring repos: %s/%s (blog), %s/%s (code)",
		envFile["GITHUB_OWNER"],
		envFile["GITHUB_REPO_WEBSITE"],
		envFile["GITHUB_OWNER"],
		envFile["GITHUB_REPO_BOT"],
	)

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

func newRouter(args router) *router {
	return &router{
		blogHandler:   args.blogHandler,
		codeHandler:   args.codeHandler,
		repoWebsite:   args.repoWebsite,
		repoBot:       args.repoBot,
		webhookSecret: args.webhookSecret,
	}
}

func (router *router) HandleWebhook(writer http.ResponseWriter, request *http.Request) {
	// read entire request body
	body, err := io.ReadAll(request.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(writer, "error reading body", http.StatusBadRequest)
		return
	}

	// parse the event type of the request
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
	default:
		log.Printf("Unknown repo detected 🛸")
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
