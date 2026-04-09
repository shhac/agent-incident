# agent-incident

incident.io triage CLI for AI agents. Focused on incident response workflows — incidents, alerts, schedules, escalations — not full incident.io administration.

## Dev Workflow

```bash
make build          # Build binary
make test           # Run all tests
make vet            # Go vet
make dev ARGS="incidents list --status active"
```

## Testing

Tests use `shared.SetupMockServer()` which creates an `httptest.Server` and injects it via `shared.ClientFactory`. Tests verify:
- Correct API paths and methods
- Request headers (Authorization: Bearer)
- Request body structure
- Error classification for HTTP status codes
- Time parsing (relative, RFC3339, epoch)

## Architecture

```
cmd/agent-incident/main.go     → cli.Execute(version)
internal/cli/root.go           → cobra root command, global flags
internal/cli/<domain>/         → domain commands (Register pattern)
internal/cli/shared/           → GlobalFlags, WithClient, output helpers, timeparse
internal/api/client.go         → HTTP client, error classification
internal/api/<domain>.go       → domain-specific API methods + types
internal/config/               → multi-org config (~/.config/agent-incident/config.json)
internal/credential/           → credential storage (keychain + file fallback)
internal/errors/               → structured errors with FixableBy classification
internal/output/               → JSON/YAML/NDJSON formatting, null pruning
```

## Key Patterns

- Domain commands register via `Register(root, globals)` where globals is `func() *shared.GlobalFlags`
- Auth commands register via `Register(root)` (no globals needed)
- `shared.WithClient(g.APIKey, g.Org, g.Timeout, func(ctx, client) error)` resolves credentials and handles errors
- All errors written to stderr as structured JSON with `fixable_by` classification
- Output: NDJSON default for lists, JSON for single items, `--format` flag overrides
- List commands use compact projections by default, `--full` for complete API responses
- `ClientFactory` var in shared package enables DI for testing

## Environment Variables

- `INCIDENT_API_KEY` — direct credential auth (skips org config)
- `INCIDENT_API_URL` — override base API URL (for testing)

## API Notes

- Base URL: `https://api.incident.io`
- Auth: `Authorization: Bearer <api-key>`
- Primary API version: v2 (some endpoints use v1)
- Rate limit: 1200 req/min per key
- Pagination: cursor-based with `after` parameter
