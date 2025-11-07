package main

import (
	"fmt"
	"log"
	"net/http"

	botAi "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot_ai"
	botBlog "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot_blog"
	botCode "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot_code"
	botGithub "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot_github"
	botRouter "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot_router"
	"github.com/joho/godotenv"
)

const PORT = "8080"

func healthCheck(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("OK"))
}

func main() {
	envFile, _ := godotenv.Read(".env")

	for key, value := range envFile {
		if value == "" {
			errMessage := fmt.Sprint("Missing required environment variable ", key)
			log.Fatal(errMessage)
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

	router := botRouter.NewRouter(
		botRouter.Router{
			BlogHandler:   blogHandler,
			CodeHandler:   codeHandler,
			RepoWebsite:   envFile["GITHUB_REPO_WEBSITE"],
			RepoBot:       envFile["GITHUB_REPO_BOT"],
			WebhookSecret: envFile["GITHUB_WEBHOOK_SECRET"],
		},
	)

	http.HandleFunc("/webhook", router.HandleWebhook)
	http.HandleFunc("/health", healthCheck)

	log.Printf("AI Blog Bot starting on :%s", PORT)

	log.Printf(
		"Monitoring repos: %s/%s (blog), %s/%s (code)",
		envFile["GITHUB_OWNER"],
		envFile["GITHUB_REPO_WEBSITE"],
		envFile["GITHUB_OWNER"],
		envFile["GITHUB_REPO_BOT"],
	)

	log.Fatal(http.ListenAndServe(":"+PORT, nil))
}
