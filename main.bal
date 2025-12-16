// GitHub Webhook Notification Service
// Main entry point - routes webhook events to configured integrations
import ballerina/http;
import ballerina/log;
import ballerina/time;
import notify.discord;

// Main webhook service
service /webhook on new http:Listener(port) {

    resource function get health() returns json {
        return {
            status: "UP",
            serviceName: "GitHub Notification Service",
            version: "0.1.0",
            timestamp: time:utcNow()[0]
        };
    }

    resource function post github(http:Request req) returns http:Ok|http:BadRequest|http:Unauthorized|http:InternalServerError {
        // Get GitHub event type
        string|http:HeaderNotFoundError eventTypeResult = req.getHeader("X-GitHub-Event");
        if eventTypeResult is http:HeaderNotFoundError {
            log:printError("Missing X-GitHub-Event header");
            return http:BAD_REQUEST;
        }
        string eventType = eventTypeResult;

        // Verify webhook signature
        string|http:HeaderNotFoundError signatureResult = req.getHeader("X-Hub-Signature-256");
        if signatureResult is http:HeaderNotFoundError {
            log:printError("Missing X-Hub-Signature-256 header");
            return http:UNAUTHORIZED;
        }
        string signature = signatureResult;

        // Get request body
        byte[]|http:ClientError payload = req.getBinaryPayload();
        if payload is http:ClientError {
            log:printError("Failed to get request payload", payload);
            return http:BAD_REQUEST;
        }

        // Verify HMAC signature
        if !verifyGitHubSignature(payload, signature) {
            log:printError("Invalid webhook signature");
            return http:UNAUTHORIZED;
        }

        // Parse JSON payload
        json|error jsonPayload = req.getJsonPayload();
        if jsonPayload is error {
            log:printError("Failed to parse JSON payload", jsonPayload);
            return http:BAD_REQUEST;
        }

        log:printInfo("Received GitHub webhook", eventType = eventType);

        // Validate organization (skip for ping events)
        if eventType != "ping" && !isFromConfiguredOrg(jsonPayload) {
            json|error repoOwner = jsonPayload.repository.owner.login;
            string owner = repoOwner is error ? "unknown" : repoOwner.toString();
            log:printInfo("Ignoring event from different organization", 
                eventOrg = owner, 
                configuredOrg = githubOrganization);
            return http:OK;
        }

        // Route to all configured integrations
        // Discord integration
        error? discordResult = discord:handleGitHubEvent(discordWebhookId, discordWebhookToken, eventType, jsonPayload);
        if discordResult is error {
            log:printError("Discord notification failed", discordResult);
        }

        // Future integrations can be added here:
        // error? slackResult = slack:handleGitHubEvent(slackWebhookUrl, eventType, jsonPayload);
        // error? telegramResult = telegram:handleGitHubEvent(telegramBotToken, telegramChatId, eventType, jsonPayload);

        return http:OK;
    }
}

public function main() returns error? {
    log:printInfo("Starting GitHub Notification Service on port: " + port.toString());
    log:printInfo("Monitoring organization: github.com/" + githubOrganization);
    log:printInfo("Webhook endpoint: POST /webhook/github");
    log:printInfo("Health check: GET /webhook/health");
    log:printInfo("Active integrations: Discord");
}
