# Notify - GitHub Organization Notification System

A Ballerina-based multi-platform notification system that receives GitHub organization events via webhooks and distributes them across Discord, Telegram, WhatsApp, and Slack.

## 🚀 Features

- ✅ GitHub webhook integration with signature validation
- ✅ Multi-platform notifications (Discord, Telegram, WhatsApp, Slack)
- ✅ Asynchronous queue-based processing
- ✅ Event filtering (by type, branch, author)
- ✅ Retry logic with exponential backoff
- ✅ Rich message formatting for each platform
- ✅ Health check and status endpoints

## 📋 Prerequisites

- Ballerina 2201.10.0 or later
- GitHub organization with webhook access
- Platform accounts:
  - Discord webhook URL
  - Telegram bot token
  - Twilio account (for WhatsApp)
  - Slack webhook URL (optional)

## 🛠️ Installation

1. Clone the repository:
```bash
cd notify
```

2. Copy the sample config and fill in your credentials:
```bash
cp Config.toml.sample Config.toml
# Edit Config.toml with your credentials
```

3. Build the project:
```bash
bal build
```

## ⚙️ Configuration

Edit `Config.toml` with your credentials:

```toml
[github]
webhookSecret = "your-github-webhook-secret"
organizations = ["your-org"]

[platforms.discord]
enabled = true
webhookUrl = "https://discord.com/api/webhooks/..."

[platforms.telegram]
enabled = true
botToken = "YOUR_BOT_TOKEN"
chatIds = ["-1001234567890"]
```

## 🚦 Running

```bash
bal run
```

The service will start on `http://localhost:8080`

## 🔗 GitHub Webhook Setup

1. Go to your GitHub organization settings
2. Navigate to Webhooks
3. Add webhook with:
   - Payload URL: `https://your-domain.com/webhook`
   - Content type: `application/json`
   - Secret: Same as in Config.toml
   - Events: Select events you want to receive

## 📡 API Endpoints

- `GET /webhook/health` - Health check
- `GET /webhook/status` - System status and queue size
- `POST /webhook` - GitHub webhook receiver

## 🧪 Testing

```bash
bal test
```

## 📦 Project Structure

```
notify/
├── main.bal                  # Entry point
├── modules/
│   ├── github/              # GitHub webhook handling
│   ├── notifiers/           # Platform integrations
│   ├── queue/               # Message queue
│   ├── config/              # Configuration
│   └── utils/               # Utilities
├── resources/               # Templates and schemas
└── tests/                   # Test files
```

## 🎯 Supported Events

- Push events
- Pull Request events (opened, merged, closed)
- Issue events
- Release events
- Star events
- And more...

## 🔐 Security

- Webhook signature validation using HMAC-SHA256
- Secrets stored in Config.toml (gitignored)
- Environment variable support for production

## 📝 License

MIT

## 🤝 Contributing

Contributions welcome! Please open an issue or PR.

## 👤 Author

**Pruthivithejan**

---

Made with ❤️ using Ballerina
