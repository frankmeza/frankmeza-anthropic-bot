# AI Bot Usage Guide

## Overview

Your AI bot monitors two repositories and responds to GitHub issues:

- **`frankmeza/frankmeza`** → Generates blog posts in markdown
- **`frankmeza/frankmeza-anthropic-bot`** → Generates code changes in Go

## How It Works

1. **Create an issue** with the right format
2. **Bot reacts** with 👍 to acknowledge
3. **Bot creates a branch** and opens a PR
4. **Review and comment** on the PR for changes
5. **Bot updates** based on your feedback

---

## Blog Posts (frankmeza/frankmeza)

### Issue Format

**Title must contain:** "Blog post:" or "blog post:"

```markdown
Title: Blog post: [Your Topic]

Body:
Describe what you want the blog post to cover. Be as detailed as you like.
You can mention specific points, tone, examples you want included.

Optional:
Tags: golang, htmx, web-development
```

### Examples

#### Simple Request
```markdown
Title: Blog post: Why Go and HTMX work well together

Body:
I want to explain why I chose Go and HTMX for my personal website.
Cover the developer experience, performance benefits, and simplicity.
Include some practical code examples.
```

#### Detailed Request
```markdown
Title: Blog post: Building a GitHub webhook handler in Go

Body:
Write a tutorial-style post about creating a GitHub webhook handler.

Key points to cover:
- Setting up the webhook in GitHub
- Validating webhook signatures
- Parsing different event types
- Best practices for error handling
- Testing webhooks locally with ngrok

Include working code examples and keep the tone casual but informative.

Tags: golang, github, webhooks, tutorial
```

### PR Interaction

After the bot creates a PR, you can comment to request changes:

**Content changes:**
- "Can you add more code examples?"
- "Make this more technical"
- "Add a section about error handling"
- "The tone is too casual, make it more professional"
- "Expand on the performance implications"

**Publishing:**
- "Ready to publish!" → Moves from drafts/ to posts/
- "Move back to draft" → Moves from posts/ to drafts/

---

## Code Changes (frankmeza-anthropic-bot)

### Issue Format

**Title must contain one of:**
- "Code:" or "code:"
- "Add feature"
- "Refactor"
- "Implement"

```markdown
Title: Code: [What you want to add/change]

Body:
Describe the code change you want. Be specific about:
- What functionality to add
- Where it should go (optional)
- Any specific patterns or approaches to use

Optional:
File: pkg/bot-ai/client.go
Path: pkg/bot-code/helpers.go
```

### Examples

#### Adding a New Feature
```markdown
Title: Code: Add rate limiting to AI client

Body:
Implement rate limiting for the Anthropic API calls to prevent hitting API limits.

Requirements:
- Use a token bucket or sliding window approach
- Limit to 50 requests per minute
- Add exponential backoff when rate limited
- Include clear error messages

File: pkg/bot-ai/client.go
```

#### Refactoring
```markdown
Title: Refactor: Extract common error handling

Body:
Create a shared error handling function that can be used across all handlers.
The function should:
- Log errors with context
- Format user-friendly error messages
- Track error counts for monitoring

Target location would be a new file: pkg/common/errors.go
```

#### Adding Tests
```markdown
Title: Code: Add tests for blog post parser

Body:
Create unit tests for the ParseIssueForRequest function in pkg/bot-blog/blog.go

Test cases to cover:
- Basic blog post request parsing
- Extracting tags from issue body
- Handling malformed input
- Edge cases with special characters in titles
```

#### Improving Existing Code
```markdown
Title: Code: Add retry logic to GitHub API calls

Body:
Add automatic retry with exponential backoff for all GitHub API operations.

Details:
- Retry up to 3 times
- Use exponential backoff (1s, 2s, 4s)
- Only retry on transient errors (5xx, rate limits)
- Log each retry attempt

File: pkg/bot-github/client.go
```

### PR Interaction

Comment on code PRs to request modifications:

**Code improvements:**
- "Can you add error handling for X?"
- "This function is too long, please split it up"
- "Add comments explaining the algorithm"
- "Use a more descriptive variable name"
- "Add input validation"

**Refactoring:**
- "Extract this logic into a separate function"
- "Make this more idiomatic Go"
- "Simplify the error handling"

---

## Tips for Best Results

### Be Specific
❌ "Add logging"
✅ "Add structured logging to the HandleWebhook function that logs the event type, repo name, and processing time"

### Provide Context
❌ "Fix the bug"
✅ "When an issue has no body, the parser crashes. Add nil checking and return a helpful error message"

### Include Examples
```markdown
For example, if someone comments "make this faster", the bot should:
1. Profile the current code
2. Identify bottlenecks
3. Apply optimizations
4. Add benchmarks
```

### Iterate in PRs
Don't try to get everything perfect in the initial issue. The bot learns from your PR comments, so:

1. Start with a basic request
2. Review the generated code/content
3. Request specific improvements in PR comments
4. Iterate until it's right

### File Paths
When you specify a file path, the bot will try to create/modify that exact file. If you don't specify a path, it will infer based on the request:

- "handler" → `pkg/bot-code/handlers.go`
- "client" → `pkg/bot-code/client.go`
- "test" → `pkg/bot-code/code_test.go`

---

## What the Bot Does Well

✅ Generating boilerplate code
✅ Creating new functions with standard patterns
✅ Writing documentation and comments
✅ Implementing well-defined algorithms
✅ Refactoring for readability
✅ Adding error handling
✅ Creating tests for existing functions
✅ Writing blog posts with code examples

## What to Review Carefully

⚠️ Complex business logic
⚠️ Security-sensitive code
⚠️ Performance-critical sections
⚠️ API contract changes
⚠️ Database migrations
⚠️ Deployment configurations

Always review the generated code/content before merging!

---

## Troubleshooting

### Bot doesn't respond to issue
- Check title format (must contain trigger words)
- Verify webhook is configured correctly
- Check bot logs for errors

### Generated code doesn't compile
- Comment on the PR with the error message
- Ask bot to fix specific compilation issues
- May need to provide more context about dependencies

### Output is cut off
- Token limit reached (currently 5000 tokens)
- Ask bot to "complete the code" or "finish the function"
- Or split into multiple smaller issues

### Wrong file location
- Specify the exact path in the issue body
- Use `File:` or `Path:` prefix

---

## Example Workflow

### Self-Improving Bot Example

```markdown
Title: Code: Add validation for issue titles

Body:
Add a validation function that checks if issue titles are well-formed
before processing them.

Validation rules:
- Title should not be empty
- For blog posts: must start with "Blog post:"
- For code changes: must contain "Code:", "Refactor", "Add feature", or "Implement"
- Return clear error messages for invalid titles

File: pkg/bot-blog/handlers.go
```

1. Bot creates PR with validation function
2. You review and comment: "Can you also add logging for invalid titles?"
3. Bot updates PR with logging
4. You comment: "Extract the validation rules into constants"
5. Bot refactors with constants
6. Merge! 🎉

Now your bot is slightly better at handling issues!

---

## Advanced: Chaining Changes

You can create multiple issues that build on each other:

1. **Issue 1:** "Code: Add logging interface"
2. Wait for PR, review, merge
3. **Issue 2:** "Code: Use logging interface in handlers"
4. Wait for PR, review, merge
5. **Issue 3:** "Code: Add tests for logging"

This way the bot builds features incrementally.

---

## Have Fun!

The bot learns from your feedback, so the more you use it and provide clear corrections in PR comments, the better it gets at understanding your preferences and coding style.

Remember: You're not just using an AI tool, you're improving it through every interaction! 🤖✨
