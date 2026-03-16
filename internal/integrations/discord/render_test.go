package discord

import "testing"

func TestRenderGitHubEventBuildsPingNotification(t *testing.T) {
	payload, err := RenderGitHubEvent("ping", map[string]any{
		"zen": "keep it logically awesome",
	})
	if err != nil {
		t.Fatalf("expected ping render to succeed, got error: %v", err)
	}

	if payload.Username != "GitHub Notify" {
		t.Fatalf("expected username GitHub Notify, got %q", payload.Username)
	}

	if len(payload.Embeds) != 1 {
		t.Fatalf("expected 1 embed, got %d", len(payload.Embeds))
	}

	if payload.Embeds[0].Title != "GitHub Webhook Connected" {
		t.Fatalf("expected ping title, got %q", payload.Embeds[0].Title)
	}
}

func TestRenderGitHubEventBuildsPullRequestNotification(t *testing.T) {
	payload, err := RenderGitHubEvent("pull_request", map[string]any{
		"action": "opened",
		"pull_request": map[string]any{
			"number":     float64(42),
			"title":      "Add webhook migration",
			"html_url":   "https://github.com/org/repo/pull/42",
			"state":      "open",
			"body":       "This changes everything.",
			"user":       map[string]any{"login": "octocat", "html_url": "https://github.com/octocat", "avatar_url": "https://avatars.example/octocat.png"},
			"merged":     false,
			"created_at": "2026-03-17T00:00:00Z",
		},
		"repository": map[string]any{
			"full_name": "org/repo",
			"html_url":  "https://github.com/org/repo",
		},
	})
	if err != nil {
		t.Fatalf("expected pull request render to succeed, got error: %v", err)
	}

	embed := payload.Embeds[0]
	if embed.Title != "New Pull Request: #42 Add webhook migration" {
		t.Fatalf("unexpected pull request title: %q", embed.Title)
	}

	if embed.Color != ColorBlue {
		t.Fatalf("expected pull request color %d, got %d", ColorBlue, embed.Color)
	}
}

func TestRenderGitHubEventSkipsUnsupportedPullRequestAction(t *testing.T) {
	_, err := RenderGitHubEvent("pull_request", map[string]any{
		"action": "edited",
		"pull_request": map[string]any{
			"number":   float64(42),
			"title":    "Add webhook migration",
			"html_url": "https://github.com/org/repo/pull/42",
			"state":    "open",
			"user":     map[string]any{"login": "octocat"},
		},
		"repository": map[string]any{
			"full_name": "org/repo",
			"html_url":  "https://github.com/org/repo",
		},
	})
	if err == nil {
		t.Fatal("expected unsupported pull request action to return an error")
	}
}

func TestRenderGitHubEventBuildsPushNotification(t *testing.T) {
	payload, err := RenderGitHubEvent("push", map[string]any{
		"ref":     "refs/heads/main",
		"compare": "https://github.com/org/repo/compare/old...new",
		"commits": []any{
			map[string]any{
				"id":      "abcdef1234567890",
				"message": "Add Discord renderer",
				"url":     "https://github.com/org/repo/commit/abcdef1",
				"author":  map[string]any{"name": "Octo Cat"},
			},
		},
		"repository": map[string]any{
			"full_name": "org/repo",
			"html_url":  "https://github.com/org/repo",
		},
		"sender": map[string]any{
			"login":      "octocat",
			"html_url":   "https://github.com/octocat",
			"avatar_url": "https://avatars.example/octocat.png",
		},
	})
	if err != nil {
		t.Fatalf("expected push render to succeed, got error: %v", err)
	}

	embed := payload.Embeds[0]
	if embed.Title != "1 new commit to main" {
		t.Fatalf("unexpected push title: %q", embed.Title)
	}

	if embed.Footer == nil || embed.Footer.Text != "GitHub Push" {
		t.Fatalf("expected GitHub Push footer, got %#v", embed.Footer)
	}
}

func TestRenderGitHubEventBuildsStarNotification(t *testing.T) {
	payload, err := RenderGitHubEvent("star", map[string]any{
		"action": "created",
		"repository": map[string]any{
			"name":     "notify",
			"html_url": "https://github.com/org/notify",
		},
		"sender": map[string]any{
			"login":      "octocat",
			"html_url":   "https://github.com/octocat",
			"avatar_url": "https://avatars.example/octocat.png",
		},
	})
	if err != nil {
		t.Fatalf("expected star render to succeed, got error: %v", err)
	}

	if payload.Embeds[0].Title != "New star on notify" {
		t.Fatalf("unexpected star title: %q", payload.Embeds[0].Title)
	}
}

func TestRenderGitHubEventBuildsIssueNotification(t *testing.T) {
	payload, err := RenderGitHubEvent("issues", map[string]any{
		"action": "closed",
		"issue": map[string]any{
			"number":     float64(7),
			"title":      "Fix webhook edge case",
			"html_url":   "https://github.com/org/repo/issues/7",
			"state":      "closed",
			"body":       "Closed issue body",
			"user":       map[string]any{"login": "octocat", "html_url": "https://github.com/octocat", "avatar_url": "https://avatars.example/octocat.png"},
			"created_at": "2026-03-17T00:00:00Z",
		},
		"repository": map[string]any{
			"full_name": "org/repo",
			"html_url":  "https://github.com/org/repo",
		},
	})
	if err != nil {
		t.Fatalf("expected issue render to succeed, got error: %v", err)
	}

	if payload.Embeds[0].Title != "Issue Closed: #7 Fix webhook edge case" {
		t.Fatalf("unexpected issue title: %q", payload.Embeds[0].Title)
	}
}

func TestRenderGitHubEventBuildsReleaseNotification(t *testing.T) {
	payload, err := RenderGitHubEvent("release", map[string]any{
		"action": "published",
		"release": map[string]any{
			"tag_name":   "v1.0.0",
			"name":       "Version 1",
			"html_url":   "https://github.com/org/repo/releases/tag/v1.0.0",
			"body":       "Release notes",
			"prerelease": false,
			"author":     map[string]any{"login": "octocat", "html_url": "https://github.com/octocat", "avatar_url": "https://avatars.example/octocat.png"},
		},
		"repository": map[string]any{
			"full_name": "org/repo",
			"html_url":  "https://github.com/org/repo",
		},
	})
	if err != nil {
		t.Fatalf("expected release render to succeed, got error: %v", err)
	}

	if payload.Embeds[0].Title != "New Release: Version 1" {
		t.Fatalf("unexpected release title: %q", payload.Embeds[0].Title)
	}
}

func TestRenderGitHubEventBuildsCreateNotification(t *testing.T) {
	payload, err := RenderGitHubEvent("create", map[string]any{
		"ref":      "feature-branch",
		"ref_type": "branch",
		"repository": map[string]any{
			"full_name": "org/repo",
			"html_url":  "https://github.com/org/repo",
		},
		"sender": map[string]any{
			"login":      "octocat",
			"html_url":   "https://github.com/octocat",
			"avatar_url": "https://avatars.example/octocat.png",
		},
	})
	if err != nil {
		t.Fatalf("expected create render to succeed, got error: %v", err)
	}

	if payload.Embeds[0].Title != "New branch created: feature-branch" {
		t.Fatalf("unexpected create title: %q", payload.Embeds[0].Title)
	}
}

func TestRenderGitHubEventBuildsDeleteNotification(t *testing.T) {
	payload, err := RenderGitHubEvent("delete", map[string]any{
		"ref":      "old-branch",
		"ref_type": "branch",
		"repository": map[string]any{
			"full_name": "org/repo",
			"html_url":  "https://github.com/org/repo",
		},
		"sender": map[string]any{
			"login":      "octocat",
			"html_url":   "https://github.com/octocat",
			"avatar_url": "https://avatars.example/octocat.png",
		},
	})
	if err != nil {
		t.Fatalf("expected delete render to succeed, got error: %v", err)
	}

	if payload.Embeds[0].Title != "branch deleted: old-branch" {
		t.Fatalf("unexpected delete title: %q", payload.Embeds[0].Title)
	}
}

func TestRenderGitHubEventBuildsForkNotification(t *testing.T) {
	payload, err := RenderGitHubEvent("fork", map[string]any{
		"forkee": map[string]any{
			"full_name": "octocat/repo-fork",
			"html_url":  "https://github.com/octocat/repo-fork",
		},
		"repository": map[string]any{
			"full_name": "org/repo",
			"html_url":  "https://github.com/org/repo",
		},
		"sender": map[string]any{
			"login":      "octocat",
			"html_url":   "https://github.com/octocat",
			"avatar_url": "https://avatars.example/octocat.png",
		},
	})
	if err != nil {
		t.Fatalf("expected fork render to succeed, got error: %v", err)
	}

	if payload.Embeds[0].Title != "Repository forked" {
		t.Fatalf("unexpected fork title: %q", payload.Embeds[0].Title)
	}
}
