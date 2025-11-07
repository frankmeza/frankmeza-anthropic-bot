package botrouter

import (
	"bytes"
	"io"
	"log"
	"net/http"

	botBlog "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot_blog"
	botCode "github.com/frankmeza/frankmeza-anthropic-bot/pkg/bot_code"
	"github.com/google/go-github/v57/github"
)

// router handles routing webhooks to the appropriate handler
type Router struct {
	BlogHandler   *botBlog.Handler
	CodeHandler   *botCode.Handler
	RepoWebsite   string
	RepoBot       string
	WebhookSecret string // Add this
}

func NewRouter(args Router) *Router {
	return &Router{
		BlogHandler:   args.BlogHandler,
		CodeHandler:   args.CodeHandler,
		RepoWebsite:   args.RepoWebsite,
		RepoBot:       args.RepoBot,
		WebhookSecret: args.WebhookSecret,
	}
}

func (router *Router) HandleWebhook(writer http.ResponseWriter, request *http.Request) {
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
	case contains(repoName, router.RepoWebsite):
		log.Printf("Routing to blog handler")
		router.BlogHandler.HandleWebhook(writer, request)

	case contains(repoName, router.RepoBot):
		log.Printf("Routing to code handler")
		router.CodeHandler.HandleWebhook(writer, request)

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
