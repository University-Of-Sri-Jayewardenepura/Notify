# Go Rewrite Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace the current Ballerina GitHub-to-Discord webhook service with a Go implementation that serves as `v1` of an extensible GitHub organization operations microservice, using a layered package structure and staged in-place cutover commits.

**Architecture:** Build a Go service under `cmd/notify` and `internal/...`, keeping `v1` feature scope limited to the current webhook notifier while separating workflow orchestration from tool integrations. Migrate the HTTP edge and webhook verification first, introduce reusable workflow and integration boundaries, then move Discord delivery, switch deployment assets, and finally remove the Ballerina implementation.

**Tech Stack:** Go, standard library HTTP server, `go test`, Docker, docker-compose

---

### Task 1: Scaffold the Go module and package layout

**Files:**
- Create: `go.mod`
- Create: `cmd/notify/main.go`
- Create: `internal/config/config.go`
- Create: `internal/httpapi/router.go`
- Create: `internal/httpapi/health.go`
- Create: `internal/service/service.go`
- Create: `internal/github/types.go`
- Create: `internal/workflows/notify/workflow.go`
- Create: `internal/integrations/discord/client.go`
- Test: `internal/httpapi/health_test.go`

**Step 1: Create the Go module**

Run: `go mod init github.com/pruthivithejan/notify`
Expected: `go.mod` is created at the repo root

**Step 2: Add the minimal entrypoint**

Create `cmd/notify/main.go` with a `main()` that:
- loads config
- constructs the router
- starts the HTTP server on the configured port

**Step 3: Add empty package stubs**

Create the initial files listed above with package declarations and minimal exported types/functions so the project compiles.

**Step 4: Add a minimal health route**

Implement `GET /webhook/health` in `internal/httpapi/health.go` returning JSON with:
- `status`
- `serviceName`
- `version`
- `timestamp`

**Step 5: Verify compilation**

Run: `go test ./...`
Expected: PASS with zero or minimal placeholder tests

**Step 6: Commit**

Run:
```bash
git add go.mod cmd/notify internal
git commit -m "feat: scaffold Go service structure"
```

### Task 2: Add config loading and validation

**Files:**
- Modify: `internal/config/config.go`
- Create: `internal/config/config_test.go`
- Modify: `cmd/notify/main.go`
- Modify: `.env.example`
- Modify: `Config.toml.example`

**Step 1: Write the failing config tests**

Create tests for:
- default port behavior
- required GitHub org and webhook secret
- required Discord webhook ID and token
- optional handling of currently unused values

Run: `go test ./internal/config -v`
Expected: FAIL because config loading is not implemented

**Step 2: Implement config loading**

Load configuration from environment variables. Use a clear mapping such as:
- `PORT`
- `GITHUB_ORGANIZATION`
- `GITHUB_WEBHOOK_SECRET`
- `DISCORD_WEBHOOK_ID`
- `DISCORD_WEBHOOK_TOKEN`

Keep the examples updated so deployment changes stay obvious.

**Step 3: Wire config into startup**

Update `cmd/notify/main.go` so startup fails fast on missing required config.

**Step 4: Verify tests**

Run: `go test ./internal/config -v`
Expected: PASS

**Step 5: Commit**

Run:
```bash
git add internal/config cmd/notify/main.go .env.example Config.toml.example
git commit -m "feat: add Go configuration loading"
```

### Task 3: Migrate GitHub webhook verification and request handling

**Files:**
- Create: `internal/github/signature.go`
- Create: `internal/github/signature_test.go`
- Create: `internal/httpapi/github_handler.go`
- Modify: `internal/httpapi/router.go`
- Modify: `internal/service/service.go`

**Step 1: Write failing signature tests**

Cover:
- valid `sha256=` signatures
- invalid signatures
- malformed headers
- case-insensitive hex comparison

Run: `go test ./internal/github -v`
Expected: FAIL

**Step 2: Implement signature verification**

Use HMAC-SHA256 over the raw request body and compare with the header after removing the `sha256=` prefix.

**Step 3: Implement the webhook handler**

In `internal/httpapi/github_handler.go`, add `POST /webhook/github` handling that:
- validates `X-GitHub-Event`
- validates `X-Hub-Signature-256`
- reads the raw body
- verifies the signature
- decodes JSON into a generic payload
- delegates to the service layer

**Step 4: Verify handler behavior**

Create or expand tests so:
- missing event header returns `400`
- missing signature returns `401`
- invalid JSON returns `400`
- invalid signature returns `401`

Run: `go test ./internal/httpapi ./internal/github -v`
Expected: PASS

**Step 5: Commit**

Run:
```bash
git add internal/github internal/httpapi internal/service
git commit -m "feat: migrate webhook verification flow"
```

### Task 4: Add organization filtering and event dispatch

**Files:**
- Create: `internal/github/filter.go`
- Create: `internal/github/filter_test.go`
- Create: `internal/domain/events.go`
- Modify: `internal/service/service.go`
- Modify: `internal/workflows/notify/workflow.go`

**Step 1: Write failing org filter tests**

Cover:
- matching org owner
- non-matching org owner
- malformed payload without repository owner
- `ping` bypass behavior at the service layer

Run: `go test ./internal/github ./internal/service -v`
Expected: FAIL

**Step 2: Implement org filtering**

Move org ownership checks into `internal/github/filter.go`.

**Step 3: Add service-level dispatch shape**

Add a service entrypoint that:
- skips org filtering for `ping`
- ignores non-matching org events
- routes eligible events into the notification workflow

**Step 4: Verify tests**

Run: `go test ./internal/github ./internal/service -v`
Expected: PASS

**Step 5: Commit**

Run:
```bash
git add internal/github internal/domain internal/service
git commit -m "feat: add event filtering and dispatch"
```

### Task 5: Add workflow and integration boundaries

**Files:**
- Modify: `internal/service/service.go`
- Modify: `internal/workflows/notify/workflow.go`
- Create: `internal/workflows/contracts.go`
- Create: `internal/integrations/contracts.go`
- Create: `internal/workflows/notify/workflow_test.go`

**Step 1: Write failing workflow dispatch tests**

Cover:
- service invokes the notification workflow for supported events
- workflow can depend on an abstract delivery integration
- ignored events do not invoke the workflow

Run: `go test ./internal/workflows/... ./internal/service -v`
Expected: FAIL

**Step 2: Implement workflow and integration contracts**

Define narrow interfaces so:
- workflows consume GitHub events and decide actions
- integrations perform tool-specific delivery
- future org-ops workflows can be added without changing HTTP or GitHub parsing layers

**Step 3: Verify tests**

Run: `go test ./internal/workflows/... ./internal/service -v`
Expected: PASS

**Step 4: Commit**

Run:
```bash
git add internal/service internal/workflows internal/integrations
git commit -m "feat: add workflow and integration boundaries"
```

### Task 6: Migrate the Discord client and payload types

**Files:**
- Modify: `internal/integrations/discord/client.go`
- Create: `internal/integrations/discord/types.go`
- Create: `internal/integrations/discord/client_test.go`

**Step 1: Write failing Discord client tests**

Cover:
- webhook URL/path construction
- successful 2xx response handling
- non-2xx response handling

Run: `go test ./internal/integrations/discord -v`
Expected: FAIL

**Step 2: Implement Discord payload and client types**

Add Go structs for:
- webhook payload
- embed
- field
- author
- footer

Implement a client that posts JSON to Discord.

**Step 3: Verify tests**

Run: `go test ./internal/integrations/discord -v`
Expected: PASS

**Step 4: Commit**

Run:
```bash
git add internal/integrations/discord
git commit -m "feat: add Discord webhook client"
```

### Task 7: Migrate GitHub event renderers

**Files:**
- Create: `internal/github/payloads.go`
- Create: `internal/integrations/discord/render.go`
- Create: `internal/integrations/discord/render_test.go`
- Modify: `internal/workflows/notify/workflow.go`

**Step 1: Write failing renderer tests using fixtures**

Add focused tests for:
- `pull_request`
- `issues`
- `push`
- `release`
- `create`
- `delete`
- `fork`
- `star`
- `ping`

Run: `go test ./internal/integrations/discord ./internal/workflows/... -v`
Expected: FAIL

**Step 2: Port the render logic incrementally**

Translate the Ballerina event formatting into Go, preserving the main notification semantics and event filtering rules.

**Step 3: Verify tests**

Run: `go test ./internal/integrations/discord ./internal/workflows/... -v`
Expected: PASS

**Step 4: Commit**

Run:
```bash
git add internal/github internal/integrations/discord internal/workflows
git commit -m "feat: migrate GitHub event rendering"
```

### Task 8: Add fixture-based parity tests and sample payloads

**Files:**
- Create: `internal/testdata/*.json`
- Create: `internal/httpapi/github_handler_test.go`
- Modify: `internal/integrations/discord/render_test.go`

**Step 1: Add representative GitHub webhook fixtures**

Create JSON fixtures for the supported event set using realistic payload shapes.

**Step 2: Add handler-level tests**

Verify end-to-end request handling with:
- headers
- raw body
- signature validation
- response code behavior

**Step 3: Run full Go test suite**

Run: `go test ./...`
Expected: PASS

**Step 4: Commit**

Run:
```bash
git add internal/testdata internal/httpapi internal/integrations/discord
git commit -m "test: add webhook parity fixtures"
```

### Task 9: Switch deployment assets to Go

**Files:**
- Modify: `Dockerfile`
- Modify: `docker-compose.yml`
- Modify: `README.md`
- Modify: `.dockerignore`
- Modify: `.env.example`

**Step 1: Replace the Ballerina image build**

Update `Dockerfile` to build and run the Go binary.

**Step 2: Update compose and env docs**

Switch compose and examples to the Go environment variable names and binary.

**Step 3: Update README**

Replace Ballerina setup and run instructions with Go build/run instructions.

**Step 4: Verify container-related changes**

Run: `go test ./...`
Expected: PASS

**Step 5: Commit**

Run:
```bash
git add Dockerfile docker-compose.yml README.md .dockerignore .env.example
git commit -m "chore: switch runtime and docs to Go"
```

### Task 10: Remove the Ballerina implementation

**Files:**
- Delete: `main.bal`
- Delete: `config.bal`
- Delete: `utils.bal`
- Delete: `modules/discord/discord.bal`
- Delete: `Ballerina.toml`
- Delete: `Dependencies.toml`
- Modify: `.gitignore`

**Step 1: Verify the Go app is the only active implementation**

Run: `go test ./...`
Expected: PASS

**Step 2: Remove obsolete Ballerina files**

Delete the Ballerina runtime and source files once the Go implementation fully replaces them.

**Step 3: Clean up ignores**

Remove ignore entries that only existed for Ballerina-generated artifacts if they are no longer needed.

**Step 4: Verify repository state**

Run: `go test ./...`
Expected: PASS

**Step 5: Commit**

Run:
```bash
git add -A
git commit -m "refactor: remove Ballerina implementation"
```

### Task 11: Final verification and cleanup

**Files:**
- Modify: `docs/plans/2026-03-17-go-rewrite.md`
- Modify: `README.md`

**Step 1: Run final verification**

Run:
- `go test ./...`
- `git status --short`

Expected:
- tests pass
- no unexpected uncommitted files remain

**Step 2: Update plan notes if implementation diverged**

Adjust the plan doc only if final file paths or commit boundaries changed materially.

**Step 3: Commit any final cleanup**

Run:
```bash
git add README.md docs/plans/2026-03-17-go-rewrite.md
git commit -m "chore: finalize Go migration cleanup"
```
