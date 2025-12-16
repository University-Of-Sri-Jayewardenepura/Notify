// Shared configuration variables for the notification service
// These are reusable across all integration modules

// Server configuration
configurable int port = 8080;

// GitHub configuration - shared across all integrations
configurable string githubOrganization = ?;
configurable string githubWebhookSecret = ?;
configurable string githubToken = "";

// Discord configuration
configurable string discordWebhookId = ?;
configurable string discordWebhookToken = ?;

// Future integrations can add their configurable variables here:
// configurable string slackWebhookUrl = "";
// configurable string telegramBotToken = "";
// configurable string telegramChatId = "";
// configurable string whatsappApiKey = "";
