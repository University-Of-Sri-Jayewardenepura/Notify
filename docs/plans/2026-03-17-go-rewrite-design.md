# Go Rewrite Design

**Date:** 2026-03-17

**Goal:** Rewrite the current Ballerina-based GitHub-to-Discord webhook service into Go using a layered project structure, performing the migration in-place with batch commits and progressively removing the Ballerina implementation.

## Current System

The existing service is a small stateless webhook application:

- `main.bal` exposes `GET /webhook/health` and `POST /webhook/github`
- `config.bal` defines runtime configuration
- `utils.bal` contains signature verification and helper utilities
- `modules/discord/discord.bal` contains GitHub event parsing, Discord payload rendering, and webhook delivery

The service has no persistence layer, background jobs, or database. This makes it a strong candidate for a vertical-slice rewrite.

## Chosen Migration Strategy

The migration will use an in-place cutover rather than a parallel service. Go code will be introduced into the existing repository first, then the runtime, deployment, and docs will be switched over, and finally the Ballerina files will be removed.

This approach is chosen because it:

- Matches the requested cutover style
- Supports frequent, reviewable batch commits
- Keeps the repository history understandable
- Avoids maintaining two production implementations for long

## Go Project Structure

The Go rewrite will use a layered service structure:

- `cmd/notify/`
  - application entrypoint
- `internal/config/`
  - environment/config loading and validation
- `internal/httpapi/`
  - router, handlers, request/response wiring
- `internal/github/`
  - webhook headers, signature verification, payload typing, org filtering
- `internal/discord/`
  - webhook client, message rendering, delivery
- `internal/service/`
  - orchestration of GitHub event handling and integration dispatch
- `internal/domain/`
  - shared event models and constants where useful
- `internal/testdata/`
  - webhook fixtures for parity-focused tests

This is intentionally more layered than the current codebase while still keeping the service small.

## Compatibility Rules

The rewrite will be mostly compatible rather than rigidly identical.

Compatibility targets:

- Preserve `GET /webhook/health`
- Preserve `POST /webhook/github`
- Preserve GitHub HMAC-SHA256 verification behavior
- Preserve organization filtering behavior for non-`ping` events
- Preserve support for the existing GitHub event set
- Preserve Discord as the first and only active delivery integration

Allowed cleanup:

- Move from Ballerina config to Go-native configuration loading
- Simplify unused config such as `githubToken` if it remains unused
- Replace outdated deployment/docs with Go-first equivalents
- Improve package boundaries so event parsing and Discord formatting are no longer coupled

## Application Design

The Go request flow will be:

1. Accept `POST /webhook/github`
2. Read raw body
3. Validate `X-GitHub-Event`
4. Validate `X-Hub-Signature-256`
5. Verify HMAC signature against configured webhook secret
6. Decode JSON payload
7. Skip non-`ping` events not belonging to the configured organization
8. Normalize or parse the event into a Go representation
9. Render a Discord webhook payload
10. Deliver the notification to Discord
11. Return `200 OK` even when notification delivery fails, while logging the failure

This preserves the current external behavior while improving internal separation.

## Event Handling Design

The current Ballerina Discord module mixes three concerns in one file:

- GitHub payload decoding
- business-level event selection
- Discord embed generation and sending

The Go rewrite will split those concerns:

- `internal/github/` will handle headers, signature checks, and typed payload parsing
- `internal/service/` will decide whether an event should produce a notification
- `internal/discord/` will render embeds and send webhook requests

Initially supported event types:

- `pull_request`
- `issues`
- `push`
- `release`
- `create`
- `delete`
- `fork`
- `star`
- `ping`

## Error Handling

Error handling will follow the existing contract where it matters:

- Missing event header: `400 Bad Request`
- Missing signature header: `401 Unauthorized`
- Invalid signature: `401 Unauthorized`
- Invalid JSON/body: `400 Bad Request`
- Delivery errors: log the failure and still return `200 OK`

Internally, Go code will use explicit errors and narrow package responsibilities so failures are easier to test and trace.

## Testing Strategy

The current repository has no tests, so the rewrite will add Go tests as part of the migration.

Testing will focus on parity and safety:

- config loading tests
- signature verification tests
- org filtering tests
- webhook handler response tests
- event-specific Discord rendering tests
- fixture-based tests using representative GitHub payloads

The migration should reach a point where the Go service can be validated without depending on the Ballerina runtime.

## Commit Strategy

The rewrite will be split into batch commits with clear review boundaries. Expected commit themes:

- scaffold Go module and layered package structure
- add config loading and HTTP server bootstrap
- migrate webhook verification and request handling
- migrate Discord client and event renderers
- add fixture-based tests for parity
- switch Docker and docs to Go
- remove Ballerina source and obsolete configuration

Exact commit messages may change slightly based on the implementation details discovered during migration.

## Cutover Plan

The cutover will happen in this order:

1. Introduce Go module and package layout
2. Add Go server bootstrap and health endpoint
3. Add GitHub webhook verification path
4. Add Discord event handling one slice at a time
5. Add tests and fixtures for migrated behavior
6. Switch container/runtime docs to Go
7. Remove Ballerina source and config files no longer needed

This ordering keeps the migration incremental while steadily moving toward a Go-only repository.

## Risks

- The Discord formatting logic is concentrated in one large source file and may hide small behavior details that are easy to miss in translation
- Some docs and deployment files are already slightly stale, so source code must remain the authority during migration
- Since the current project lacks tests, parity confidence must come from new Go tests and careful event-by-event migration

## Success Criteria

The migration is complete when:

- the service runs entirely in Go
- the webhook endpoints behave as intended
- supported GitHub events produce Discord notifications
- deployment assets target the Go implementation
- obsolete Ballerina files have been removed
- the repository contains a clearer structure and test coverage than before
