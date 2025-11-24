import ballerina/http;
import ballerina/log;
import ballerina/time;

// Simplified main service - compiles without errors
configurable int port = 8080;
configurable string webhookSecret = "change-me";

service /webhook on new http:Listener(port) {
    
    resource function get health() returns json {
        return {
            status: "UP",
            serviceName: "GitHub Notification System",
            version: "0.1.0",
            timestamp: time:utcNow()[0]
        };
    }

    resource function post .(http:Request req) returns string|http:BadRequest|http:Unauthorized {
        
        string|http:HeaderNotFoundError eventTypeHeader = req.getHeader("X-GitHub-Event");
        string|http:HeaderNotFoundError signatureHeader = req.getHeader("X-Hub-Signature-256");
        
        if eventTypeHeader is http:HeaderNotFoundError || signatureHeader is http:HeaderNotFoundError {
            log:printError("Missing required GitHub headers");
            http:BadRequest badReq = http:BAD_REQUEST;
            return badReq;
        }

        log:printInfo("Received GitHub webhook");
        return "Webhook received successfully";
    }
}

public function main() returns error? {
    log:printInfo("Starting GitHub Notification System on port: " + port.toString());
}
