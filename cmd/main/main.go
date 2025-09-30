package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	botAi "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot-ai"
	botBlog "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot-blog"
	botGithub "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot-github"
)

func main() {
	// Load environment variables
	aiApiKey := os.Getenv("AI_API_KEY")
	githubToken := os.Getenv("GITHUB_TOKEN")
	owner := os.Getenv("GITHUB_OWNER")
	repo := os.Getenv("GITHUB_REPO")
	webhookSecret := os.Getenv("GITHUB_WEBHOOK_SECRET")

	// Validate required environment variables
	if githubToken == "" || aiApiKey == "" || owner == "" || repo == "" || webhookSecret == "" {
		fmt.Printf("isMissingAiApiKey %v", aiApiKey == "")
		fmt.Printf("isMissingGithubToken %v", githubToken == "")
		fmt.Printf("isMissingOwner %v", owner == "")
		fmt.Printf("isMissingRepo %v", repo == "")
		fmt.Printf("isMissingWebhookSecret %v", webhookSecret == "")

		log.Fatal("Missing required environment variables")
	}

	// Initialize clients
	aiClient := botAi.NewClient(aiApiKey)
	githubClient := botGithub.NewClient(githubToken)

	// Initialize blog handler
	blogHandler := botBlog.NewHandler(githubClient, aiClient, owner, repo, webhookSecret)

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
