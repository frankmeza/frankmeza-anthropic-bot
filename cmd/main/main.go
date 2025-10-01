package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	botai "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot-ai"
	botblog "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot-blog"
	botcode "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot-code"
	botgithub "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot-github"
	"github.com/google/go-github/v57/github"
)

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

	githubClient := botgithub.NewClient(githubToken)
	aiClient := botai.NewClient(aiAPIKey)

	blogHandler := botblog.NewHandler(githubClient, aiClient, owner, repoWebsite, webhookSecret)
	codeHandler := botcode.NewHandler(githubClient, aiClient, owner, repoBot)

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
	blogHandler   *botblog.Handler
	codeHandler   *botcode.Handler
	repoWebsite   string
	repoBot       string
	webhookSecret string // Add this
}

func newRouter(blogHandler *botblog.Handler, codeHandler *botcode.Handler, repoWebsite, repoBot, webhookSecret string) *router {
	return &router{
		blogHandler:   blogHandler,
		codeHandler:   codeHandler,
		repoWebsite:   repoWebsite,
		repoBot:       repoBot,
		webhookSecret: webhookSecret, // Add this
	}
}
func (rt *router) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Try to get repo from header first
	repoName := r.Header.Get("X-GitHub-Repository")

	// If header not found, parse the payload to determine repo
	if repoName == "" {
		log.Printf("X-GitHub-Repository header not found, parsing payload...")

		payload, err := github.ValidatePayload(r, []byte(rt.webhookSecret))
		if err != nil {
			log.Printf("Webhook validation failed: %v", err)
			http.Error(w, "validation failed", http.StatusUnauthorized)
			return
		}

		event, err := github.ParseWebHook(github.WebHookType(r), payload)
		if err != nil {
			log.Printf("Webhook parsing failed: %v", err)
			http.Error(w, "parsing failed", http.StatusBadRequest)
			return
		}

		// Extract repo name from the event
		fmt.Println("event %v", event)
		switch e := event.(type) {
		case *github.IssuesEvent:
			repoName = *e.Repo.FullName
		case *github.PullRequestReviewCommentEvent:
			repoName = *e.Repo.FullName
		}

		log.Printf("Detected repo from payload: %s", repoName)
	}

	switch {
	case contains(repoName, rt.repoWebsite):
		log.Printf("Routing to blog handler for repo: %s", repoName)
		rt.blogHandler.HandleWebhook(w, r)

	case contains(repoName, rt.repoBot):
		log.Printf("Routing to code handler for repo: %s", repoName)
		rt.codeHandler.HandleWebhook(w, r)

	default:
		log.Printf("Unknown repository: %s", repoName)
		w.WriteHeader(http.StatusOK)
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || s[len(s)-len(substr):] == substr)
}
