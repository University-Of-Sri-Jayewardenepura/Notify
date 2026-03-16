package discord

import (
	"fmt"
	"strings"
	"time"
)

const (
	ColorGreen  = 3066993
	ColorBlue   = 3447003
	ColorRed    = 15158332
	ColorPurple = 10181046
	ColorYellow = 16776960
	ColorOrange = 15105570
)

func RenderGitHubEvent(eventType string, payload map[string]any) (WebhookPayload, error) {
	switch eventType {
	case "pull_request":
		return renderPullRequest(payload)
	case "issues":
		return renderIssue(payload)
	case "push":
		return renderPush(payload)
	case "release":
		return renderRelease(payload)
	case "create":
		return renderCreate(payload)
	case "delete":
		return renderDelete(payload)
	case "fork":
		return renderFork(payload)
	case "star":
		return renderStar(payload)
	case "ping":
		return WebhookPayload{
			Username: "GitHub Notify",
			Embeds: []Embed{
				{
					Title:       "GitHub Webhook Connected",
					Description: "Webhook has been successfully configured and is now active.",
					Color:       ColorGreen,
					Timestamp:   currentTimestamp(),
				},
			},
		}, nil
	default:
		return WebhookPayload{}, fmt.Errorf("unhandled event type: %s", eventType)
	}
}

func renderPullRequest(payload map[string]any) (WebhookPayload, error) {
	action := stringValue(payload, "action")
	if action != "opened" && action != "closed" && action != "reopened" && action != "ready_for_review" {
		return WebhookPayload{}, fmt.Errorf("skipping PR action: %s", action)
	}

	pr := mapValue(payload, "pull_request")
	repo := mapValue(payload, "repository")
	user := mapValue(pr, "user")

	description := stringValue(pr, "body")
	if description == "" {
		description = "No description provided."
	}
	description = truncateText(description, 200)

	title := "Pull Request Ready for Review"
	color := ColorPurple
	merged := boolValue(pr, "merged")

	switch {
	case action == "closed" && merged:
		title = "Pull Request Merged"
		color = ColorGreen
	case action == "closed":
		title = "Pull Request Closed"
		color = ColorRed
	case action == "opened":
		title = "New Pull Request"
		color = ColorBlue
	case action == "reopened":
		title = "Pull Request Reopened"
		color = ColorYellow
	}

	return newPayload(Embed{
		Title:       fmt.Sprintf("%s: #%d %s", title, intValue(pr, "number"), stringValue(pr, "title")),
		Description: description,
		URL:         stringValue(pr, "html_url"),
		Color:       color,
		Timestamp:   currentTimestamp(),
		Author: &EmbedAuthor{
			Name:    stringValue(user, "login"),
			URL:     stringValue(user, "html_url"),
			IconURL: stringValue(user, "avatar_url"),
		},
		Fields: []EmbedField{
			inlineField("Repository", markdownLink(stringValue(repo, "full_name"), stringValue(repo, "html_url"))),
			inlineField("State", stringValue(pr, "state")),
		},
		Footer: &EmbedFooter{Text: "GitHub Pull Request"},
	}), nil
}

func renderIssue(payload map[string]any) (WebhookPayload, error) {
	action := stringValue(payload, "action")
	if action != "opened" && action != "closed" && action != "reopened" {
		return WebhookPayload{}, fmt.Errorf("skipping issue action: %s", action)
	}

	issue := mapValue(payload, "issue")
	repo := mapValue(payload, "repository")
	user := mapValue(issue, "user")

	title := "Issue Reopened"
	color := ColorYellow
	switch action {
	case "opened":
		title = "New Issue"
		color = ColorBlue
	case "closed":
		title = "Issue Closed"
		color = ColorGreen
	}

	description := stringValue(issue, "body")
	if description == "" {
		description = "No description provided."
	}
	description = truncateText(description, 200)

	return newPayload(Embed{
		Title:       fmt.Sprintf("%s: #%d %s", title, intValue(issue, "number"), stringValue(issue, "title")),
		Description: description,
		URL:         stringValue(issue, "html_url"),
		Color:       color,
		Timestamp:   currentTimestamp(),
		Author: &EmbedAuthor{
			Name:    stringValue(user, "login"),
			URL:     stringValue(user, "html_url"),
			IconURL: stringValue(user, "avatar_url"),
		},
		Fields: []EmbedField{
			inlineField("Repository", markdownLink(stringValue(repo, "full_name"), stringValue(repo, "html_url"))),
		},
		Footer: &EmbedFooter{Text: "GitHub Issue"},
	}), nil
}

func renderPush(payload map[string]any) (WebhookPayload, error) {
	repo := mapValue(payload, "repository")
	sender := mapValue(payload, "sender")
	commits := sliceValue(payload, "commits")
	if len(commits) == 0 {
		return WebhookPayload{}, fmt.Errorf("no commits in push")
	}

	branch := extractBranchName(stringValue(payload, "ref"))
	commitWord := "commit"
	if len(commits) > 1 {
		commitWord = "commits"
	}

	commitLines := make([]string, 0, min(len(commits), 5))
	for _, rawCommit := range commits[:min(len(commits), 5)] {
		commit := anyMap(rawCommit)
		author := mapValue(commit, "author")
		message := firstLine(stringValue(commit, "message"))
		if len(message) > 50 {
			message = message[:47] + "..."
		}

		sha := stringValue(commit, "id")
		if len(sha) > 7 {
			sha = sha[:7]
		}

		commitLines = append(commitLines, fmt.Sprintf("[`%s`](%s) %s - %s", sha, stringValue(commit, "url"), message, stringValue(author, "name")))
	}

	description := strings.Join(commitLines, "\n")
	if len(commits) > 5 {
		description += fmt.Sprintf("\n... and %d more commits", len(commits)-5)
	}

	footer := "GitHub Push"
	if boolValue(payload, "forced") {
		footer = "Force Push"
	}

	return newPayload(Embed{
		Title:       fmt.Sprintf("%d new %s to %s", len(commits), commitWord, branch),
		Description: description,
		URL:         stringValue(payload, "compare"),
		Color:       ColorOrange,
		Timestamp:   currentTimestamp(),
		Author: &EmbedAuthor{
			Name:    stringValue(sender, "login"),
			URL:     stringValue(sender, "html_url"),
			IconURL: stringValue(sender, "avatar_url"),
		},
		Fields: []EmbedField{
			inlineField("Repository", markdownLink(stringValue(repo, "full_name"), stringValue(repo, "html_url"))),
			inlineField("Branch", branch),
		},
		Footer: &EmbedFooter{Text: footer},
	}), nil
}

func renderRelease(payload map[string]any) (WebhookPayload, error) {
	if stringValue(payload, "action") != "published" {
		return WebhookPayload{}, fmt.Errorf("skipping release action: %s", stringValue(payload, "action"))
	}

	release := mapValue(payload, "release")
	repo := mapValue(payload, "repository")
	author := mapValue(release, "author")

	description := stringValue(release, "body")
	if description == "" {
		description = "No release notes provided."
	}
	description = truncateText(description, 300)

	releaseType := "Release"
	if boolValue(release, "prerelease") {
		releaseType = "Pre-release"
	}

	name := stringValue(release, "name")
	if name == "" {
		name = stringValue(release, "tag_name")
	}

	return newPayload(Embed{
		Title:       fmt.Sprintf("New %s: %s", releaseType, name),
		Description: description,
		URL:         stringValue(release, "html_url"),
		Color:       ColorGreen,
		Timestamp:   currentTimestamp(),
		Author: &EmbedAuthor{
			Name:    stringValue(author, "login"),
			URL:     stringValue(author, "html_url"),
			IconURL: stringValue(author, "avatar_url"),
		},
		Fields: []EmbedField{
			inlineField("Repository", markdownLink(stringValue(repo, "full_name"), stringValue(repo, "html_url"))),
			inlineField("Tag", stringValue(release, "tag_name")),
		},
		Footer: &EmbedFooter{Text: "GitHub Release"},
	}), nil
}

func renderCreate(payload map[string]any) (WebhookPayload, error) {
	repo := mapValue(payload, "repository")
	sender := mapValue(payload, "sender")

	return newPayload(Embed{
		Title:     fmt.Sprintf("New %s created: %s", stringValue(payload, "ref_type"), stringValue(payload, "ref")),
		Color:     ColorBlue,
		Timestamp: currentTimestamp(),
		Author: &EmbedAuthor{
			Name:    stringValue(sender, "login"),
			URL:     stringValue(sender, "html_url"),
			IconURL: stringValue(sender, "avatar_url"),
		},
		Fields: []EmbedField{
			inlineField("Repository", markdownLink(stringValue(repo, "full_name"), stringValue(repo, "html_url"))),
			inlineField("Type", stringValue(payload, "ref_type")),
		},
		Footer: &EmbedFooter{Text: "GitHub Create"},
	}), nil
}

func renderDelete(payload map[string]any) (WebhookPayload, error) {
	repo := mapValue(payload, "repository")
	sender := mapValue(payload, "sender")

	return newPayload(Embed{
		Title:     fmt.Sprintf("%s deleted: %s", stringValue(payload, "ref_type"), stringValue(payload, "ref")),
		Color:     ColorRed,
		Timestamp: currentTimestamp(),
		Author: &EmbedAuthor{
			Name:    stringValue(sender, "login"),
			URL:     stringValue(sender, "html_url"),
			IconURL: stringValue(sender, "avatar_url"),
		},
		Fields: []EmbedField{
			inlineField("Repository", markdownLink(stringValue(repo, "full_name"), stringValue(repo, "html_url"))),
		},
		Footer: &EmbedFooter{Text: "GitHub Delete"},
	}), nil
}

func renderFork(payload map[string]any) (WebhookPayload, error) {
	repo := mapValue(payload, "repository")
	forkee := mapValue(payload, "forkee")
	sender := mapValue(payload, "sender")

	return newPayload(Embed{
		Title:       "Repository forked",
		Description: markdownLink(stringValue(forkee, "full_name"), stringValue(forkee, "html_url")),
		Color:       ColorPurple,
		Timestamp:   currentTimestamp(),
		Author: &EmbedAuthor{
			Name:    stringValue(sender, "login"),
			URL:     stringValue(sender, "html_url"),
			IconURL: stringValue(sender, "avatar_url"),
		},
		Fields: []EmbedField{
			inlineField("Original", markdownLink(stringValue(repo, "full_name"), stringValue(repo, "html_url"))),
		},
		Footer: &EmbedFooter{Text: "GitHub Fork"},
	}), nil
}

func renderStar(payload map[string]any) (WebhookPayload, error) {
	if stringValue(payload, "action") != "created" {
		return WebhookPayload{}, fmt.Errorf("skipping star action: %s", stringValue(payload, "action"))
	}

	repo := mapValue(payload, "repository")
	sender := mapValue(payload, "sender")

	return newPayload(Embed{
		Title:     fmt.Sprintf("New star on %s", stringValue(repo, "name")),
		URL:       stringValue(repo, "html_url"),
		Color:     ColorYellow,
		Timestamp: currentTimestamp(),
		Author: &EmbedAuthor{
			Name:    stringValue(sender, "login"),
			URL:     stringValue(sender, "html_url"),
			IconURL: stringValue(sender, "avatar_url"),
		},
		Footer: &EmbedFooter{Text: "GitHub Star"},
	}), nil
}

func newPayload(embed Embed) WebhookPayload {
	return WebhookPayload{
		Username: "GitHub Notify",
		Embeds:   []Embed{embed},
	}
}

func inlineField(name string, value string) EmbedField {
	return EmbedField{Name: name, Value: value, Inline: true}
}

func currentTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func truncateText(text string, max int) string {
	if len(text) <= max {
		return text
	}
	return text[:max-3] + "..."
}

func extractBranchName(ref string) string {
	return strings.TrimPrefix(ref, "refs/heads/")
}

func firstLine(text string) string {
	if index := strings.Index(text, "\n"); index >= 0 {
		return text[:index]
	}
	return text
}

func markdownLink(label string, url string) string {
	return fmt.Sprintf("[%s](%s)", label, url)
}

func stringValue(values map[string]any, key string) string {
	if raw, ok := values[key]; ok {
		switch value := raw.(type) {
		case string:
			return value
		}
	}
	return ""
}

func boolValue(values map[string]any, key string) bool {
	raw, ok := values[key]
	if !ok {
		return false
	}
	value, ok := raw.(bool)
	return ok && value
}

func intValue(values map[string]any, key string) int {
	raw, ok := values[key]
	if !ok {
		return 0
	}
	switch value := raw.(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	default:
		return 0
	}
}

func mapValue(values map[string]any, key string) map[string]any {
	if raw, ok := values[key]; ok {
		return anyMap(raw)
	}
	return map[string]any{}
}

func anyMap(value any) map[string]any {
	result, ok := value.(map[string]any)
	if !ok {
		return map[string]any{}
	}
	return result
}

func sliceValue(values map[string]any, key string) []any {
	raw, ok := values[key]
	if !ok {
		return nil
	}
	result, ok := raw.([]any)
	if ok {
		return result
	}
	return nil
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
