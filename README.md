# agent-incident

incident.io triage CLI for AI agents. Incidents, alerts, schedules, escalations — not full incident.io administration.

- **Token-efficient output** — NDJSON for lists, JSON for single items, YAML available. Compact and null-pruned by default. `--full` for complete API responses
- **Structured error classification** — every error includes `fixable_by: agent|human|retry` so AI agents can self-correct without parsing messages
- **Triage-focused** — only the commands you need during an investigation, not the 165+ incident.io API endpoints
- **Multi-org support** — switch between incident.io organizations with `--organization`, credentials stored in macOS Keychain
- **Self-documenting** — `agent-incident llm-help` and per-domain `agent-incident <domain> llm-help` for agent-friendly reference

### Why not the incident.io API directly?

The raw API returns deeply nested objects with many fields irrelevant to triage. `agent-incident` is built for AI agents: compact projections that fit in context windows, classified errors to stderr, cursor pagination handled automatically, and a surface area limited to what matters during incident response.

## Installation

```bash
brew install shhac/tap/agent-incident
```

### Claude Code / AI agent skill

```bash
npx skills add shhac/agent-incident
```

### Other options

Download binaries from [GitHub Releases](https://github.com/shhac/agent-incident/releases), or build from source:

```bash
go install github.com/shhac/agent-incident/cmd/agent-incident@latest
```

## Quick start

### 1. Add an organization

```bash
agent-incident auth add prod --api-key <INCIDENT_API_KEY>
agent-incident auth check
```

Or use an environment variable directly (no setup needed):

```bash
export INCIDENT_API_KEY=<key>
```

### 2. Check active incidents

```bash
# All active incidents
agent-incident incidents list --status active

# Full details for a specific incident
agent-incident incidents get inc_123
```

### 3. Triage

```bash
# What alerts are firing?
agent-incident alerts list --status firing

# Who's on-call?
agent-incident schedules entries <schedule-id> --from now --to now+1h

# Escalate
agent-incident escalations create --incident inc_123 --path <path-id>
```

## Command map

```text
agent-incident
├── auth            add, check, default, list, remove
├── incidents       list, get, create, edit, updates
├── alerts          list, get, create, incidents
├── severities      list, get
├── statuses        list, get
├── users           list, get
├── roles           list, get
├── schedules       list, get, entries, override
├── escalations     list, get, create, paths list, paths get
├── actions         list, get
├── follow-ups      list, get
├── catalog         types list, types get, entries list, entries get
├── custom-fields   list, get
├── status-pages    list, incidents list, incidents create, incidents update
├── llm-help        top-level reference card
└── version
```

## Output

- **stdout** — NDJSON for list commands (one object per line), JSON for single-item commands
- **stderr** — errors as JSON with `fixable_by` classification
- **`--format json|yaml|jsonl`** — override the default for any command
- **Compact by default** — e.g. incidents show `id, name, status, severity, created_at, incident_lead`. Use `--full` for everything
- **Null-pruned** — empty/null fields stripped from output to save tokens

## Error output

All errors are written to stderr as structured JSON:

```json
{"error": "Not found: incident inc_999", "fixable_by": "agent", "hint": "Check the ID — use 'list' to see available items"}
```

| `fixable_by` | Meaning |
|---|---|
| `agent` | Bad request — the agent should fix its parameters and retry |
| `human` | Auth or permissions — the agent should stop and ask the human |
| `retry` | Transient — rate limit or server error, wait and retry |

## Multi-org support

```bash
# Add multiple organizations
agent-incident auth add prod --api-key <key>
agent-incident auth add staging --api-key <key>

# Query a specific org
agent-incident incidents list --status active --organization staging

# Set a default
agent-incident auth default prod
```

## Time formats

All `--since` / `--from` / `--to` flags accept:

- **Relative** — `now-15m`, `now-1h`, `now-7d`, `now+1h`
- **RFC3339** — `2024-01-15T10:00:00Z`
- **Unix epoch** — `1705312800`

## API key scopes

When creating an API key at `https://app.incident.io/~/settings/api-keys`, select these roles based on what you need:

| Role | Required for | Required? |
|---|---|---|
| `viewer` | All read operations (incidents, alerts, severities, statuses, users, roles, custom-fields, catalog) | **Yes** |
| `incident_creator` | `incidents create` | If creating incidents |
| `incident_editor` | `incidents edit` | If editing incidents |
| `catalog_viewer` | `catalog types/entries list/get` | If browsing catalog |
| `schedules_reader` | `schedules list/get/entries` | If checking on-call |
| `schedule_overrides_editor` | `schedules override` | If overriding on-call |
| `escalation_creator` | `escalations create` | If escalating incidents |
| `status_page_publisher` | `status-pages incidents create/update` | If updating status pages |

**Minimum for read-only triage:** `viewer` + `schedules_reader` + `catalog_viewer`

## Environment variables

| Variable | Purpose |
|---|---|
| `INCIDENT_API_KEY` | Direct credential auth (skips org config) |
| `INCIDENT_API_URL` | Override base API URL (for testing) |

## Development

```bash
make build          # Build binary
make test           # Run all tests
make vet            # Go vet
make dev ARGS="incidents list --status active"
```

## License

MIT
