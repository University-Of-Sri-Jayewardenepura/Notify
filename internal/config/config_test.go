package config

import (
	"strings"
	"testing"
)

func TestLoadUsesDefaultPortWhenUnset(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("GITHUB_ORGANIZATION", "acme")
	t.Setenv("GITHUB_WEBHOOK_SECRET", "secret")
	t.Setenv("DISCORD_WEBHOOK_ID", "123")
	t.Setenv("DISCORD_WEBHOOK_TOKEN", "token")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected config to load, got error: %v", err)
	}

	if cfg.Port != "8080" {
		t.Fatalf("expected default port 8080, got %q", cfg.Port)
	}
}

func TestLoadReturnsErrorWhenRequiredValuesMissing(t *testing.T) {
	testCases := []struct {
		name        string
		envKey      string
		expectedErr string
	}{
		{
			name:        "organization",
			envKey:      "GITHUB_ORGANIZATION",
			expectedErr: "GITHUB_ORGANIZATION is required",
		},
		{
			name:        "webhook secret",
			envKey:      "GITHUB_WEBHOOK_SECRET",
			expectedErr: "GITHUB_WEBHOOK_SECRET is required",
		},
		{
			name:        "discord webhook id",
			envKey:      "DISCORD_WEBHOOK_ID",
			expectedErr: "DISCORD_WEBHOOK_ID is required",
		},
		{
			name:        "discord webhook token",
			envKey:      "DISCORD_WEBHOOK_TOKEN",
			expectedErr: "DISCORD_WEBHOOK_TOKEN is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("PORT", "9090")
			t.Setenv("GITHUB_ORGANIZATION", "acme")
			t.Setenv("GITHUB_WEBHOOK_SECRET", "secret")
			t.Setenv("DISCORD_WEBHOOK_ID", "123")
			t.Setenv("DISCORD_WEBHOOK_TOKEN", "token")
			t.Setenv(tc.envKey, "")

			_, err := Load()
			if err == nil {
				t.Fatal("expected config load to fail when required values are missing")
			}

			if !strings.Contains(err.Error(), tc.expectedErr) {
				t.Fatalf("expected error %q, got %q", tc.expectedErr, err.Error())
			}
		})
	}
}

func TestLoadAllowsOptionalGitHubTokenToBeEmpty(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("GITHUB_ORGANIZATION", "acme")
	t.Setenv("GITHUB_WEBHOOK_SECRET", "secret")
	t.Setenv("DISCORD_WEBHOOK_ID", "123")
	t.Setenv("DISCORD_WEBHOOK_TOKEN", "token")
	t.Setenv("GITHUB_TOKEN", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected config to load, got error: %v", err)
	}

	if cfg.GitHubToken != "" {
		t.Fatalf("expected optional GitHub token to be empty, got %q", cfg.GitHubToken)
	}
}
