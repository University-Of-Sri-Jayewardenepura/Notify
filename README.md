# Notify - GitHub to Discord Notification Service

A Ballerina-based notification service that receives GitHub organization events via webhooks and sends rich notifications to Discord.

## Features

- GitHub webhook integration with HMAC-SHA256 signature validation
- Organization-wide monitoring (all repos in your org)
- Rich Discord embed notifications with colors and formatting
- Support for multiple GitHub event types:
  - Pull Requests (opened, merged, closed, ready for review)
  - Issues (opened, closed, reopened)
  - Push events (with commit details)
  - Releases (published)
  - Branch/Tag creation and deletion
  - Repository forks
  - Stars
- Health check endpoint

## Prerequisites

- Ballerina 2201.10.0 or later (`brew install ballerina` on macOS)
- Discord server with webhook permissions
- GitHub organization with admin access

## Quick Start

1. **Configure the service:**
```bash
cp Config.toml.example Config.toml
```

2. **Edit `Config.toml` with your values:**
```toml
port = 8080

# Discord - get these from your webhook URL
discordWebhookId = "123456789012345678"
discordWebhookToken = "your-webhook-token"

# GitHub organization to monitor (github.com/YOUR_ORG)
githubOrganization = "your-org-name"

# Secret for webhook validation (you create this)
githubWebhookSecret = "your-super-secret-key"
```

3. **Build and run:**
```bash
bal build
bal run
```

The service starts on `http://localhost:8080`

## Configuration

### Discord Webhook Setup

1. Go to your Discord server
2. Right-click the channel where you want notifications
3. Select **Edit Channel** > **Integrations** > **Webhooks**
4. Click **New Webhook** and copy the URL
5. Extract the ID and token from the URL:
   ```
   https://discord.com/api/webhooks/WEBHOOK_ID/WEBHOOK_TOKEN
                                    ^^^^^^^^^^  ^^^^^^^^^^^^^
   ```
6. Add these to your `Config.toml`

### GitHub Organization Webhook Setup

1. Go to your GitHub organization: `github.com/YOUR_ORG`
2. Navigate to **Settings** > **Webhooks** > **Add webhook**
3. Configure the webhook:

| Field | Value |
|-------|-------|
| **Payload URL** | `https://your-server.com/webhook/github` |
| **Content type** | `application/json` |
| **Secret** | Same value as `githubWebhookSecret` in Config.toml |

4. Under **Which events would you like to trigger this webhook?**, select **Let me select individual events** and choose:
   - Pull requests
   - Issues
   - Pushes
   - Releases
   - Create (for branch/tag creation)
   - Delete (for branch/tag deletion)
   - Forks
   - Stars
   - (or select "Send me everything")

5. Ensure **Active** is checked
6. Click **Add webhook**

GitHub will send a `ping` event to verify the connection.

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/webhook/health` | GET | Health check - returns service status |
| `/webhook/github` | POST | GitHub webhook receiver |

## Event Notifications

| Event | Color | Description |
|-------|-------|-------------|
| PR Opened | Blue | New pull request created |
| PR Merged | Green | Pull request merged |
| PR Closed | Red | Pull request closed without merge |
| Issue Opened | Blue | New issue created |
| Issue Closed | Green | Issue resolved |
| Push | Orange | Commits pushed (shows up to 5) |
| Release | Green | New release published |
| Branch/Tag Created | Blue | New branch or tag |
| Branch/Tag Deleted | Red | Branch or tag removed |
| Fork | Purple | Repository forked |
| Star | Yellow | New star on repository |

## Development

```bash
# Build
bal build

# Run in development
bal run

# Run the built JAR
java -jar target/bin/notify.jar
```

## Deployment

For production, you'll need to expose the service to the internet. Options:
- Use a reverse proxy (nginx, Caddy) with HTTPS
- Deploy to a cloud platform (AWS, GCP, Azure)
- Use a tunnel service (ngrok, cloudflared) for testing

## Security

- All GitHub webhooks are validated using HMAC-SHA256 signatures
- Only events from the configured organization are processed
- Secrets are stored in `Config.toml` (gitignored)
- Never commit your actual `Config.toml` with real credentials

## Project Structure

```
notify/
├── main.bal              # Main application code
├── Ballerina.toml        # Project configuration
├── Config.toml.example   # Example configuration (safe to commit)
├── Config.toml           # Your actual config (gitignored)
├── .env.example          # Environment variables reference
└── resources/
    └── templates/        # Message templates
```

## License

MIT

## Author

**Pruthivithejan**

---

Built with Ballerina
