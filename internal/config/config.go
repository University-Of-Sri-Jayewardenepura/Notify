package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port                string
	GitHubOrganization  string
	GitHubWebhookSecret string
	GitHubToken         string
	DiscordWebhookID    string
	DiscordWebhookToken string
}

func Load() (Config, error) {
	cfg := Config{
		Port:                valueOrDefault("PORT", "8080"),
		GitHubOrganization:  strings.TrimSpace(os.Getenv("GITHUB_ORGANIZATION")),
		GitHubWebhookSecret: strings.TrimSpace(os.Getenv("GITHUB_WEBHOOK_SECRET")),
		GitHubToken:         strings.TrimSpace(os.Getenv("GITHUB_TOKEN")),
		DiscordWebhookID:    strings.TrimSpace(os.Getenv("DISCORD_WEBHOOK_ID")),
		DiscordWebhookToken: strings.TrimSpace(os.Getenv("DISCORD_WEBHOOK_TOKEN")),
	}

	switch {
	case cfg.GitHubOrganization == "":
		return Config{}, fmt.Errorf("GITHUB_ORGANIZATION is required")
	case cfg.GitHubWebhookSecret == "":
		return Config{}, fmt.Errorf("GITHUB_WEBHOOK_SECRET is required")
	case cfg.DiscordWebhookID == "":
		return Config{}, fmt.Errorf("DISCORD_WEBHOOK_ID is required")
	case cfg.DiscordWebhookToken == "":
		return Config{}, fmt.Errorf("DISCORD_WEBHOOK_TOKEN is required")
	default:
		return cfg, nil
	}
}

func valueOrDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}
