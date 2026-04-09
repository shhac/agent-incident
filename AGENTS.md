# agent-incident

incident.io triage CLI for AI agents. Focused on incident response workflows — incidents, alerts, schedules, escalations — not full incident.io administration.

## Dev Workflow

```bash
make build          # Build binary
make test           # Run all tests
make vet            # Go vet
make dev ARGS="incident list --status active"
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
internal/cli/root.go           → cobra root command, global flags, oncall/ref groups
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

- Domain commands register via `Register(parent, globals)` where parent is root, oncall, or ref group
- Auth commands register via `Register(root)` (no globals needed)
- Commands use singular names: `incident`, `alert`, `schedule` (not plural)
- `shared.WithClient(g.APIKey, g.Org, g.Timeout, func(ctx, client) error)` resolves credentials and handles errors
- All errors written to stderr as structured JSON with `fixable_by` classification
- Output: NDJSON default for lists, JSON for single items, `--format` flag overrides
- List commands use compact projections by default, `--full` for complete API responses
- `ClientFactory` var in shared package enables DI for testing

## Environment Variables

- `INCIDENT_API_KEY` — direct credential auth (skips org config)
- `INCIDENT_API_URL` — override base API URL (for testing)

## Documentation Checklist

When adding or changing commands, flags, or features, always update:
- **`internal/cli/<domain>/usage.go`** — LLM help text for the domain
- **`internal/cli/usage.go`** — top-level `llm-help` reference card
- **`README.md`** — command map and examples (if user-facing structure changed)
- **`skills/agent-incident/SKILL.md`** — quick reference, discovery commands, workflow examples

## Data Hygiene

- **Never store real IDs** in docs, tests, or fixtures — use synthetic IDs (e.g. `sev-crit`, `01AAABBBCCC000000000000000`)
- **Never store real data from incident.io** in docs, tests, or fixtures — use generic examples
- To diagnose unexpected API behavior, use `--debug` which logs full HTTP request/response to stderr

## API Notes

- Base URL: `https://api.incident.io`
- Auth: `Authorization: Bearer <api-key>`
- Primary API version: v2 (some endpoints use v1 where v2 doesn't exist)
- Rate limit: 1200 req/min per key
- Pagination: cursor-based with `after` parameter

## Subagent Roles

### Implementer

Writes production code for a specific domain (API methods + CLI commands). Owns files in `internal/api/<domain>.go` and `internal/cli/<domain>/`. Follows existing patterns from sibling domains.

### Reviewer

Reviews implementation for correctness, consistency with sibling patterns, and test coverage. Does not modify files — reports findings only.

### Test Writer

Writes tests for a specific domain using `shared.SetupMockServer`. Owns `*_test.go` files. Verifies API paths, query params, request bodies, and output format.
