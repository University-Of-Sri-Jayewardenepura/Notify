# Notify

A Go microservice that receives GitHub organization webhooks and sends formatted notifications to Discord.

## Features

- GitHub webhook validation with HMAC-SHA256 signatures
- Organization-level filtering for accepted events
- Discord webhook delivery for supported GitHub events
- Health endpoint at `GET /webhook/health`
- Webhook receiver at `POST /webhook/github`

Supported events:

- Pull requests
- Issues
- Pushes
- Releases
- Create
- Delete
- Fork
- Star
- Ping

## Prerequisites

- Go `1.25.1+`
- A Discord webhook
- GitHub organization admin access for webhook setup

## Configuration

The service reads configuration from environment variables.

Required:

- `DISCORD_WEBHOOK_ID`
- `DISCORD_WEBHOOK_TOKEN`
- `GITHUB_ORGANIZATION`
- `GITHUB_WEBHOOK_SECRET`

Optional:

- `GITHUB_TOKEN`
- `PORT` defaults to `8080`

Example:

```bash
cp .env.example .env
```

## Running Locally

```bash
set -a && source .env && set +a
go run ./cmd/notify
```

The service listens on `http://localhost:8080` by default.

## Testing

```bash
set -a && source .env && set +a
go test ./...
```

## Docker

Build and run with Docker Compose:

```bash
docker compose up -d --build
```

## GitHub Webhook Setup

1. Go to your GitHub organization settings.
2. Open `Webhooks` and create a new webhook.
3. Use `https://your-server.com/webhook/github` as the payload URL.
4. Set content type to `application/json`.
5. Use the same secret value as `GITHUB_WEBHOOK_SECRET`.
6. Enable the supported event types or choose all events.

GitHub will send a `ping` event to verify the webhook.

## Discord Webhook Setup

1. Open your Discord channel settings.
2. Go to `Integrations` -> `Webhooks`.
3. Create a webhook and copy its URL.
4. Extract the webhook ID and token from:

```text
https://discord.com/api/webhooks/WEBHOOK_ID/WEBHOOK_TOKEN
```

## Security

- Webhooks are validated before payload processing.
- Events outside the configured organization are ignored.
- Local secrets stay in `.env`, which is gitignored.
