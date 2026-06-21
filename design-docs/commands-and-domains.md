# agent-incident: Command & Domain Design

## Overview

CLI tool for AI agents (and humans) to triage, query, and manage incident.io resources.
Pattern: `agent-incident <domain> <action> [flags]` — matches agent-dd, agent-sql, agent-statsig.

## Auth

- Bearer token via `INCIDENT_API_KEY` env var (or `--api-key` flag, or stored credential)
- Supports multiple organizations, each with its own API key and alias
- Global `--organization <alias>` flag on all commands to override the default
- `agent-incident auth add <alias>` — store an API key for a named organization
- `agent-incident auth check [alias]` — verify stored key by calling `GET /v1/identity` (defaults to default org)
- `agent-incident auth default <alias>` — set the default organization
- `agent-incident auth list` — show all configured organizations and which is default
- `agent-incident auth remove <alias>` — remove a stored credential
- Store in `~/.config/agent-incident/credentials.json` (matching sibling pattern)
- Resolution order: `--api-key` flag > `INCIDENT_API_KEY` env var > `--organization` flag > default org from config

## Proposed Domains & Commands

### 1. `incidents` — Core domain, most useful for triage

```
agent-incident incidents list [--status=open|closed|...] [--severity=...] [--since=...]
agent-incident incidents get <id>...
agent-incident incidents create --name "..." --severity "..." [--summary "..."]
agent-incident incidents edit <id> [--status=...] [--severity=...] [--summary=...]
agent-incident incidents updates <id>          # list incident updates/timeline
```

**Why prioritize:** This is the primary use case — an AI agent triaging or responding to an incident needs to list, inspect, and update incidents.

### 2. `alerts` — Alert inspection and creation

```
agent-incident alerts list [--status=firing|resolved] [--source=...] [--since=...]
agent-incident alerts get <id>...
agent-incident alerts create --source-id <id> --title "..." [--description "..."]
agent-incident alerts incidents                # list alerts attached to incidents
```

**Why:** Alerts feed into incidents. Understanding what triggered helps with triage.

### 3. `schedules` — On-call schedule queries

```
agent-incident schedules list
agent-incident schedules get <id>...
agent-incident schedules entries <id> [--from=...] [--to=...]   # who's on-call
agent-incident schedules override <id> --user <id> --from ... --to ...
```

**Why:** "Who's on-call right now?" is a critical triage question.

### 4. `escalations` — Escalation paths and active escalations

```
agent-incident escalations list
agent-incident escalations get <id>...
agent-incident escalations create --incident <id> --path <id>
agent-incident escalations paths list
agent-incident escalations paths get <id>...
```

**Why:** Escalating an incident is a key action during triage.

### 5. `severities` — Severity level reference

```
agent-incident severities list
agent-incident severities get <id>...
```

**Why read-only:** These are configuration lookups used to populate other commands.

### 6. `statuses` — Incident status reference

```
agent-incident statuses list
agent-incident statuses get <id>...
```

### 7. `users` — User lookup

```
agent-incident users list [--query=...]
agent-incident users get <id>...
```

**Why:** Needed to resolve user IDs for role assignments, escalations, schedule overrides.

### 8. `roles` — Incident role reference

```
agent-incident roles list
agent-incident roles get <id>...
```

### 9. `follow-ups` — Post-incident action items

```
agent-incident follow-ups list [--incident=...]
agent-incident follow-ups get <id>...
```

### 10. `actions` — In-incident tasks

```
agent-incident actions list [--incident=...]
agent-incident actions get <id>...
```

### 11. `catalog` — Service catalog queries

```
agent-incident catalog types list
agent-incident catalog types get <id>...
agent-incident catalog entries list --type <id> [--query=...]
agent-incident catalog entries get <id>...
```

**Why:** Catalog ties incidents to services/teams. Useful for understanding blast radius.

### 12. `custom-fields` — Custom field reference

```
agent-incident custom-fields list
agent-incident custom-fields get <id>...
```

### 13. `status-pages` — Status page management

```
agent-incident status-pages list
agent-incident status-pages incidents list [--page=...]
agent-incident status-pages incidents create --page <id> --name "..."
agent-incident status-pages incidents update <id> --status "..."
```

**Why:** Updating status pages is part of incident communication.

### 14. `version` — Utility

```
agent-incident version        # show CLI version
```

---

## Domains to SKIP (for now)

| Domain | Reason |
|--------|--------|
| API Keys | Admin config, not triage |
| IP Allowlists | Admin config |
| Workflows | Complex CRUD, better done in UI |
| Managed Resources | Terraform integration |
| Telemetry | Niche |
| Audit Logs | Enterprise-only, WorkOS integration |
| Webhooks | Configuration, not querying |
| Postmortem Documents | Could add later, low triage value |
| Maintenance Windows | Could add later |
| Alert Attributes/Routes/Sources | Admin config for alert pipeline |

---

## Implementation Phases

### Phase 1: Core triage loop
- `auth`, `incidents`, `alerts`, `severities`, `statuses`
- This covers the 80% use case: "what's happening, how bad is it"

### Phase 2: People & escalation
- `users`, `roles`, `schedules`, `escalations`
- Answers: "who's on-call, who should I page"

### Phase 3: Context & follow-through
- `actions`, `follow-ups`, `catalog`, `custom-fields`
- Deeper triage context and post-incident tracking

### Phase 4: Communication
- `status-pages`
- External communication during incidents

---

## Output Format

Match sibling agents — three formats via `--format` flag:

| Format | Flag | Default for | Description |
|--------|------|-------------|-------------|
| NDJSON | `--format jsonl` | Lists, gets | One JSON object per line, streamable. Most token-efficient for LLMs. |
| JSON | `--format json` | — | Pretty-printed, 2-space indent, wrapped in `{"data": [...]}` for lists; `{"data":[...],"@unresolved":[...]}` for gets. |
| YAML | `--format yaml` | — | Slightly more compact than JSON, good for human reading. |

All output prunes null fields automatically. All list commands support `--limit` and cursor-based pagination.

### Get (single + multi)

`get <id>...` takes one or more ids and returns one result per id, in input order. Default output
is NDJSON: one line per id — the record, or `{"@unresolved":{"id","reason","fixable_by","hint"?}}`
for an id that couldn't be resolved (e.g. not found / bad id). `--format json|yaml` collapses to
one `{"data":[…], "@unresolved":[…]}` envelope. A single `get <id>` is just the one-element case
(NDJSON one line by default; was pretty JSON before — pass `--format json` for the object).
Item-level misses stay on stdout and exit 0; only a command-level failure (auth, network) goes to
stderr with exit 1 and empty stdout.

### Compact vs Full output

List commands return a **compact projection** by default to minimize token usage. Use `--full` to get the complete object.

**Compact `incidents list`** (default):
```
{"id":"inc_123","name":"API latency spike","status":"active","severity":"critical","created_at":"2026-04-09T10:00:00Z","incident_lead":"alice"}
```

**Full `incidents list --full`** adds all nested fields (description, custom fields, role assignments, timestamps, attachments, etc.)

**`incidents get <id>...`** always returns the full object (no truncation) for each id.

### Fields truncated/omitted in compact mode

| Field | Compact behavior | Why |
|-------|-----------------|-----|
| `description` / `summary` | First 200 chars, suffixed with `…` | Can be multi-paragraph markdown |
| `custom_field_entries` | Omitted | Array of all custom fields, many irrelevant to triage |
| `incident_role_assignments` | Flattened to `{role_name: user_name}` | Nested user objects with full profiles |
| `timestamps` | Omitted (only `created_at` kept) | Full array of every lifecycle timestamp |
| `external_resources` / `attachments` | Omitted | Links to Slack threads, Jira tickets, etc. |

## Project Structure

```
cmd/agent-incident/main.go
internal/
  cli/           # cobra command tree
  api/           # HTTP client, auth, pagination
  config/        # credential storage
  credential/    # keychain/file credential management
  errors/        # structured error types
  output/        # table/json/template formatting
```
