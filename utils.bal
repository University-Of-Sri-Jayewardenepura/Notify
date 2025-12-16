// Utility functions shared across all integrations
import ballerina/crypto;
import ballerina/log;
import ballerina/time;

// Verify GitHub webhook signature
public function verifyGitHubSignature(byte[] payload, string signature) returns boolean {
    if !signature.startsWith("sha256=") {
        return false;
    }

    string expectedSignature = signature.substring(7);
    byte[] secretBytes = githubWebhookSecret.toBytes();
    byte[]|crypto:Error hmacResult = crypto:hmacSha256(payload, secretBytes);
    if hmacResult is crypto:Error {
        log:printError("HMAC computation failed", hmacResult);
        return false;
    }
    string computedSignature = hmacResult.toBase16();

    return computedSignature.equalsIgnoreCaseAscii(expectedSignature);
}

// Get current timestamp in ISO format
public function getCurrentTimestamp() returns string {
    time:Utc now = time:utcNow();
    return time:utcToString(now);
}

// Validate that the event is from the configured organization
public function isFromConfiguredOrg(json payload) returns boolean {
    json|error repoOwner = payload.repository.owner.login;
    if repoOwner is error {
        return false;
    }
    string owner = repoOwner.toString();
    return owner.equalsIgnoreCaseAscii(githubOrganization);
}

// Truncate text to specified length with ellipsis
public function truncateText(string text, int maxLength) returns string {
    if text.length() > maxLength {
        return text.substring(0, maxLength - 3) + "...";
    }
    return text;
}

// Extract branch name from git ref (refs/heads/main -> main)
public function extractBranchName(string ref) returns string {
    if ref.startsWith("refs/heads/") {
        return ref.substring(11);
    }
    return ref;
}
