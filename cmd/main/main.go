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

// Helper to mask sensitive tokens
func maskToken(token string) string {
	if len(token) < 8 {
		return "***"
	}
	return token[:4] + "..." + token[len(token)-4:]
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

	githubClient := botGithub.NewClient(githubToken)
	aiClient := botAi.NewClient(aiAPIKey)

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
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

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

func newRouter(blogHandler *botBlog.Handler, codeHandler *botCode.Handler, repoWebsite, repoBot, webhookSecret string) *router {
	return &router{
		blogHandler:   blogHandler,
		codeHandler:   codeHandler,
		repoWebsite:   repoWebsite,
		repoBot:       repoBot,
		webhookSecret: webhookSecret, // Add this
	}
}

func (rt *router) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Read the body once
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "error reading body", http.StatusBadRequest)
		return
	}

	// Parse to get repo name (no validation yet)
	event, err := github.ParseWebHook(github.WebHookType(r), body)
	if err != nil {
		log.Printf("Webhook parsing failed: %v", err)
		http.Error(w, "parsing failed", http.StatusBadRequest)
		return
	}

	var repoName string
	switch e := event.(type) {
	case *github.IssuesEvent:
		repoName = *e.Repo.FullName
	case *github.PullRequestReviewCommentEvent:
		repoName = *e.Repo.FullName
	}

	log.Printf("Detected repo: %s", repoName)

	// Recreate the request body for the handler
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	switch {
	case contains(repoName, rt.repoWebsite):
		log.Printf("Routing to blog handler")
		rt.blogHandler.HandleWebhook(w, r)

	case contains(repoName, rt.repoBot):
		log.Printf("Routing to code handler")
		rt.codeHandler.HandleWebhook(w, r)

	default:
		log.Printf("Unknown repository: %s", repoName)
		w.WriteHeader(http.StatusOK)
	}
}

// func (rt *router) HandleWebhook(w http.ResponseWriter, r *http.Request) {
// 	// Try to get repo from header first
// 	repoName := r.Header.Get("X-GitHub-Repository")

// 	// If header not found, parse the payload to determine repo
// 	if repoName == "" {
// 		log.Printf("X-GitHub-Repository header not found, parsing payload...")

// 		payload, err := github.ValidatePayload(r, []byte(rt.webhookSecret))
// 		if err != nil {
// 			log.Printf("Webhook validation failed: %v", err)
// 			http.Error(w, "validation failed", http.StatusUnauthorized)
// 			return
// 		}

// 		event, err := github.ParseWebHook(github.WebHookType(r), payload)
// 		if err != nil {
// 			log.Printf("Webhook parsing failed: %v", err)
// 			http.Error(w, "parsing failed", http.StatusBadRequest)
// 			return
// 		}

// 		fmt.Printf("event %T", event)
// 		// Extract repo name from the event
// 		switch e := event.(type) {
// 		case *github.IssuesEvent:
// 			repoName = *e.Repo.FullName
// 		case *github.IssueCommentEvent:
// 			repoName = *e.Repo.FullName
// 		case *github.PullRequestReviewCommentEvent:
// 			repoName = *e.Repo.FullName
// 		default:
// 			repoName = "unknown"

// 			log.Printf("Detected repo from payload: %s", repoName)
// 		}

// 		switch {
// 		case contains(repoName, rt.repoWebsite) || repoName == rt.repoWebsite:
// 			log.Printf("Routing to blog handler for repo: %s", repoName)
// 			rt.blogHandler.HandleWebhook(w, r)

// 		case contains(repoName, rt.repoBot) || repoName == rt.repoBot:
// 			log.Printf("Routing to code handler for repo: %s", repoName)
// 			rt.codeHandler.HandleWebhook(w, r)

// 		default:
// 			log.Printf("Unknown repository: %s", repoName)
// 			w.WriteHeader(http.StatusOK)
// 		}
// 	}
// }

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || s[len(s)-len(substr):] == substr)
}
