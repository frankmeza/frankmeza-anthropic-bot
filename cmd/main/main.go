package main

import (
	"log"
	"net/http"
	"os"

	botai "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot-ai"
	botblog "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot-blog"
	botgithub "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot-github"
)

func main() {
	// Load environment variables
	githubToken := os.Getenv("GITHUB_TOKEN")
	aiAPIKey := os.Getenv("AI_API_KEY")
	owner := os.Getenv("GITHUB_OWNER")
	repo := os.Getenv("GITHUB_REPO")
	webhookSecret := os.Getenv("GITHUB_WEBHOOK_SECRET")

	// Validate required environment variables
	if githubToken == "" || aiAPIKey == "" || owner == "" || repo == "" || webhookSecret == "" {
		log.Fatal("Missing required environment variables")
	}

	// Initialize clients
	aiClient := botai.NewClient(aiAPIKey)
	githubClient := botgithub.NewClient(githubToken)

	// Initialize blog handler
	blogHandler := botblog.NewHandler(githubClient, aiClient, owner, repo, webhookSecret)

	// Set up HTTP routes
	http.HandleFunc("/webhook", blogHandler.HandleWebhook)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("AI Blog Bot starting on :%s", port)
	log.Printf("Monitoring repo: %s/%s", owner, repo)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
