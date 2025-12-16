// Discord integration module for GitHub webhook notifications
import ballerina/http;
import ballerina/log;
import ballerina/time;

// Discord API base URL
const string DISCORD_API_BASE = "https://discord.com/api/webhooks";

// Discord webhook client
final http:Client discordClient = check new (DISCORD_API_BASE);

// Discord embed colors (decimal values)
public const int COLOR_GREEN = 3066993;   // Success/Merged
public const int COLOR_BLUE = 3447003;    // Info/Opened
public const int COLOR_RED = 15158332;    // Closed/Failed
public const int COLOR_PURPLE = 10181046; // Review
public const int COLOR_YELLOW = 16776960; // Warning/Pending
public const int COLOR_ORANGE = 15105570; // Push

// Discord Embed types
type EmbedField record {
    string name;
    string value;
    boolean inline = true;
};

type EmbedAuthor record {
    string name;
    string url?;
    string icon_url?;
};

type EmbedFooter record {
    string text;
    string icon_url?;
};

type Embed record {
    string title?;
    string description?;
    string url?;
    int color?;
    string timestamp?;
    EmbedAuthor author?;
    EmbedFooter footer?;
    EmbedField[] fields?;
};

type DiscordWebhookPayload record {
    string content?;
    Embed[] embeds?;
    string username?;
    string avatar_url?;
};

// GitHub types (local to discord module to avoid cyclic imports)
type GitHubUser record {
    string login;
    string avatar_url;
    string html_url;
};

type GitHubRepository record {
    string name;
    string full_name;
    string html_url;
    GitHubUser owner;
};

type GitHubPullRequest record {
    int number;
    string title;
    string html_url;
    string state;
    string body?;
    boolean merged?;
    GitHubUser user;
    string created_at?;
    string merged_at?;
};

type GitHubIssue record {
    int number;
    string title;
    string html_url;
    string state;
    string body?;
    GitHubUser user;
};

type GitHubRelease record {
    string tag_name;
    string name;
    string html_url;
    string body?;
    boolean prerelease;
    boolean draft;
    GitHubUser author;
};

type GitHubCommit record {
    string id;
    string message;
    string url;
    record {
        string name;
        string email;
    } author;
};

type PullRequestPayload record {
    string action;
    GitHubPullRequest pull_request;
    GitHubRepository repository;
    GitHubUser sender;
};

type IssuesPayload record {
    string action;
    GitHubIssue issue;
    GitHubRepository repository;
    GitHubUser sender;
};

type PushPayload record {
    string ref;
    string compare;
    GitHubCommit[] commits;
    GitHubRepository repository;
    GitHubUser sender;
    boolean forced?;
};

type ReleasePayload record {
    string action;
    GitHubRelease release;
    GitHubRepository repository;
    GitHubUser sender;
};

type CreatePayload record {
    string ref;
    string ref_type;
    GitHubRepository repository;
    GitHubUser sender;
};

type DeletePayload record {
    string ref;
    string ref_type;
    GitHubRepository repository;
    GitHubUser sender;
};

type ForkPayload record {
    GitHubRepository forkee;
    GitHubRepository repository;
    GitHubUser sender;
};

type StarPayload record {
    string action;
    GitHubRepository repository;
    GitHubUser sender;
};

// Utility functions
function getCurrentTimestamp() returns string {
    time:Utc now = time:utcNow();
    return time:utcToString(now);
}

function truncateText(string text, int maxLength) returns string {
    if text.length() > maxLength {
        return text.substring(0, maxLength - 3) + "...";
    }
    return text;
}

function extractBranchName(string ref) returns string {
    if ref.startsWith("refs/heads/") {
        return ref.substring(11);
    }
    return ref;
}

// Send notification to Discord webhook
function sendNotification(string webhookId, string webhookToken, DiscordWebhookPayload payload) returns error? {
    string webhookPath = "/" + webhookId + "/" + webhookToken;
    http:Response response = check discordClient->post(webhookPath, payload);

    if response.statusCode >= 400 {
        string|http:ClientError body = response.getTextPayload();
        string errorMsg = body is string ? body : "Unknown error";
        log:printError("Discord API error", statusCode = response.statusCode, body = errorMsg);
        return error("Discord API error: " + response.statusCode.toString());
    }

    log:printInfo("Discord notification sent successfully");
}

// Handle GitHub events and send Discord notifications
public function handleGitHubEvent(string webhookId, string webhookToken, string eventType, json payload) returns error? {
    DiscordWebhookPayload? discordPayload = ();

    match eventType {
        "pull_request" => {
            discordPayload = check handlePullRequest(payload);
        }
        "issues" => {
            discordPayload = check handleIssue(payload);
        }
        "push" => {
            discordPayload = check handlePush(payload);
        }
        "release" => {
            discordPayload = check handleRelease(payload);
        }
        "create" => {
            discordPayload = check handleCreate(payload);
        }
        "delete" => {
            discordPayload = check handleDelete(payload);
        }
        "fork" => {
            discordPayload = check handleFork(payload);
        }
        "star" => {
            discordPayload = check handleStar(payload);
        }
        "ping" => {
            log:printInfo("Received ping event - webhook configured successfully");
            discordPayload = {
                embeds: [
                    {
                        title: "GitHub Webhook Connected",
                        description: "Webhook has been successfully configured and is now active.",
                        color: COLOR_GREEN,
                        timestamp: getCurrentTimestamp()
                    }
                ],
                username: "GitHub Notify"
            };
        }
        _ => {
            log:printInfo("Unhandled event type for Discord", eventType = eventType);
        }
    }

    if discordPayload is DiscordWebhookPayload {
        check sendNotification(webhookId, webhookToken, discordPayload);
    }
}

// Handle Pull Request events
function handlePullRequest(json payload) returns DiscordWebhookPayload|error {
    PullRequestPayload prPayload = check payload.cloneWithType();
    string action = prPayload.action;
    GitHubPullRequest pr = prPayload.pull_request;
    GitHubRepository repo = prPayload.repository;

    // Filter to important actions only
    if action != "opened" && action != "closed" && action != "reopened" && action != "ready_for_review" {
        return error("Skipping PR action: " + action);
    }

    string title;
    int color;
    string description = pr.body ?: "No description provided.";

    if action == "closed" && (pr.merged ?: false) {
        title = "Pull Request Merged";
        color = COLOR_GREEN;
    } else if action == "closed" {
        title = "Pull Request Closed";
        color = COLOR_RED;
    } else if action == "opened" {
        title = "New Pull Request";
        color = COLOR_BLUE;
    } else if action == "reopened" {
        title = "Pull Request Reopened";
        color = COLOR_YELLOW;
    } else {
        title = "Pull Request Ready for Review";
        color = COLOR_PURPLE;
    }

    description = truncateText(description, 200);

    string embedTitle = title + ": #" + pr.number.toString() + " " + pr.title;
    string repoLink = "[" + repo.full_name + "](" + repo.html_url + ")";

    return {
        embeds: [
            {
                title: embedTitle,
                description: description,
                url: pr.html_url,
                color: color,
                timestamp: getCurrentTimestamp(),
                author: {
                    name: pr.user.login,
                    url: pr.user.html_url,
                    icon_url: pr.user.avatar_url
                },
                fields: [
                    {name: "Repository", value: repoLink, inline: true},
                    {name: "State", value: pr.state, inline: true}
                ],
                footer: {
                    text: "GitHub Pull Request"
                }
            }
        ],
        username: "GitHub Notify"
    };
}

// Handle Issue events
function handleIssue(json payload) returns DiscordWebhookPayload|error {
    IssuesPayload issuePayload = check payload.cloneWithType();
    string action = issuePayload.action;
    GitHubIssue issue = issuePayload.issue;
    GitHubRepository repo = issuePayload.repository;

    // Filter to important actions only
    if action != "opened" && action != "closed" && action != "reopened" {
        return error("Skipping issue action: " + action);
    }

    string title;
    int color;

    if action == "opened" {
        title = "New Issue";
        color = COLOR_BLUE;
    } else if action == "closed" {
        title = "Issue Closed";
        color = COLOR_GREEN;
    } else {
        title = "Issue Reopened";
        color = COLOR_YELLOW;
    }

    string description = issue.body ?: "No description provided.";
    description = truncateText(description, 200);

    string embedTitle = title + ": #" + issue.number.toString() + " " + issue.title;
    string repoLink = "[" + repo.full_name + "](" + repo.html_url + ")";

    return {
        embeds: [
            {
                title: embedTitle,
                description: description,
                url: issue.html_url,
                color: color,
                timestamp: getCurrentTimestamp(),
                author: {
                    name: issue.user.login,
                    url: issue.user.html_url,
                    icon_url: issue.user.avatar_url
                },
                fields: [
                    {name: "Repository", value: repoLink, inline: true}
                ],
                footer: {
                    text: "GitHub Issue"
                }
            }
        ],
        username: "GitHub Notify"
    };
}

// Handle Push events
function handlePush(json payload) returns DiscordWebhookPayload|error {
    PushPayload pushPayload = check payload.cloneWithType();
    GitHubRepository repo = pushPayload.repository;
    GitHubCommit[] commits = pushPayload.commits;

    if commits.length() == 0 {
        return error("No commits in push");
    }

    string branch = extractBranchName(pushPayload.ref);

    string commitWord = commits.length() > 1 ? "commits" : "commit";
    string title = commits.length().toString() + " new " + commitWord + " to " + branch;

    // Build commit list
    string commitList = "";
    int maxCommits = 5;
    int numToShow = commits.length() < maxCommits ? commits.length() : maxCommits;
    
    foreach int i in 0 ..< numToShow {
        GitHubCommit c = commits[i];
        string shortSha = c.id.substring(0, 7);
        string message = c.message;
        int? newlineIdx = message.indexOf("\n");
        if newlineIdx is int {
            message = message.substring(0, newlineIdx);
        }
        if message.length() > 50 {
            message = message.substring(0, 47) + "...";
        }
        commitList = commitList + "[`" + shortSha + "`](" + c.url + ") " + message + " - " + c.author.name + "\n";
    }

    if commits.length() > maxCommits {
        int remaining = commits.length() - maxCommits;
        commitList = commitList + "... and " + remaining.toString() + " more commits";
    }

    string repoLink = "[" + repo.full_name + "](" + repo.html_url + ")";
    string footerText = pushPayload.forced == true ? "Force Push" : "GitHub Push";

    return {
        embeds: [
            {
                title: title,
                description: commitList,
                url: pushPayload.compare,
                color: COLOR_ORANGE,
                timestamp: getCurrentTimestamp(),
                author: {
                    name: pushPayload.sender.login,
                    url: pushPayload.sender.html_url,
                    icon_url: pushPayload.sender.avatar_url
                },
                fields: [
                    {name: "Repository", value: repoLink, inline: true},
                    {name: "Branch", value: branch, inline: true}
                ],
                footer: {
                    text: footerText
                }
            }
        ],
        username: "GitHub Notify"
    };
}

// Handle Release events
function handleRelease(json payload) returns DiscordWebhookPayload|error {
    ReleasePayload releasePayload = check payload.cloneWithType();
    string action = releasePayload.action;
    GitHubRelease release = releasePayload.release;
    GitHubRepository repo = releasePayload.repository;

    if action != "published" {
        return error("Skipping release action: " + action);
    }

    string releaseType = release.prerelease ? "Pre-release" : "Release";
    string description = release.body ?: "No release notes provided.";
    description = truncateText(description, 300);

    string releaseName = release.name != "" ? release.name : release.tag_name;
    string embedTitle = "New " + releaseType + ": " + releaseName;
    string repoLink = "[" + repo.full_name + "](" + repo.html_url + ")";

    return {
        embeds: [
            {
                title: embedTitle,
                description: description,
                url: release.html_url,
                color: COLOR_GREEN,
                timestamp: getCurrentTimestamp(),
                author: {
                    name: release.author.login,
                    url: release.author.html_url,
                    icon_url: release.author.avatar_url
                },
                fields: [
                    {name: "Repository", value: repoLink, inline: true},
                    {name: "Tag", value: release.tag_name, inline: true}
                ],
                footer: {
                    text: "GitHub Release"
                }
            }
        ],
        username: "GitHub Notify"
    };
}

// Handle Create events (branch/tag creation)
function handleCreate(json payload) returns DiscordWebhookPayload|error {
    CreatePayload createPayload = check payload.cloneWithType();
    GitHubRepository repo = createPayload.repository;

    string embedTitle = "New " + createPayload.ref_type + " created: " + createPayload.ref;
    string repoLink = "[" + repo.full_name + "](" + repo.html_url + ")";

    return {
        embeds: [
            {
                title: embedTitle,
                color: COLOR_BLUE,
                timestamp: getCurrentTimestamp(),
                author: {
                    name: createPayload.sender.login,
                    url: createPayload.sender.html_url,
                    icon_url: createPayload.sender.avatar_url
                },
                fields: [
                    {name: "Repository", value: repoLink, inline: true},
                    {name: "Type", value: createPayload.ref_type, inline: true}
                ],
                footer: {
                    text: "GitHub Create"
                }
            }
        ],
        username: "GitHub Notify"
    };
}

// Handle Delete events (branch/tag deletion)
function handleDelete(json payload) returns DiscordWebhookPayload|error {
    DeletePayload deletePayload = check payload.cloneWithType();
    GitHubRepository repo = deletePayload.repository;

    string embedTitle = deletePayload.ref_type + " deleted: " + deletePayload.ref;
    string repoLink = "[" + repo.full_name + "](" + repo.html_url + ")";

    return {
        embeds: [
            {
                title: embedTitle,
                color: COLOR_RED,
                timestamp: getCurrentTimestamp(),
                author: {
                    name: deletePayload.sender.login,
                    url: deletePayload.sender.html_url,
                    icon_url: deletePayload.sender.avatar_url
                },
                fields: [
                    {name: "Repository", value: repoLink, inline: true}
                ],
                footer: {
                    text: "GitHub Delete"
                }
            }
        ],
        username: "GitHub Notify"
    };
}

// Handle Fork events
function handleFork(json payload) returns DiscordWebhookPayload|error {
    ForkPayload forkPayload = check payload.cloneWithType();
    GitHubRepository repo = forkPayload.repository;
    GitHubRepository forkedRepo = forkPayload.forkee;

    string forkDescription = "[" + forkedRepo.full_name + "](" + forkedRepo.html_url + ")";
    string repoLink = "[" + repo.full_name + "](" + repo.html_url + ")";

    return {
        embeds: [
            {
                title: "Repository forked",
                description: forkDescription,
                color: COLOR_PURPLE,
                timestamp: getCurrentTimestamp(),
                author: {
                    name: forkPayload.sender.login,
                    url: forkPayload.sender.html_url,
                    icon_url: forkPayload.sender.avatar_url
                },
                fields: [
                    {name: "Original", value: repoLink, inline: true}
                ],
                footer: {
                    text: "GitHub Fork"
                }
            }
        ],
        username: "GitHub Notify"
    };
}

// Handle Star events
function handleStar(json payload) returns DiscordWebhookPayload|error {
    StarPayload starPayload = check payload.cloneWithType();
    string action = starPayload.action;
    GitHubRepository repo = starPayload.repository;

    if action != "created" {
        return error("Skipping star action: " + action);
    }

    string embedTitle = "New star on " + repo.name;

    return {
        embeds: [
            {
                title: embedTitle,
                url: repo.html_url,
                color: COLOR_YELLOW,
                timestamp: getCurrentTimestamp(),
                author: {
                    name: starPayload.sender.login,
                    url: starPayload.sender.html_url,
                    icon_url: starPayload.sender.avatar_url
                },
                footer: {
                    text: "GitHub Star"
                }
            }
        ],
        username: "GitHub Notify"
    };
}
