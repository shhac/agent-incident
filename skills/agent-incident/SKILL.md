---
name: agent-incident
description: Triage and manage incident.io incidents, alerts, schedules, escalations, and status pages. Use when the user asks about active incidents, who's on-call, alert status, escalating to responders, incident severity, follow-up actions, or status page updates.
triggers:
  - incident
  - on-call
  - escalation
  - alert
  - severity
  - status page
  - schedule
  - pager
  - follow-up
allowed-tools: Bash Read Grep Glob
---

# agent-incident — incident.io Triage CLI

Query and manage incident.io incidents, alerts, schedules, escalations, and status pages. Triage and response workflows only — not full incident.io administration.

## When to Use

- Checking what incidents are active and their severity
- Finding out who's on-call right now
- Investigating which alerts are firing
- Escalating an incident to the right responders
- Updating status pages during an incident
- Reviewing follow-up actions after resolution

## Process

### Incident response workflow

1. **What's happening?** `incidents list --status active` to see active incidents
2. **How bad is it?** `incidents get <id>` for full details including severity and timeline
3. **What triggered it?** `alerts list --status firing` to see current alerts
4. **Who's on-call?** `schedules entries <schedule-id> --from now --to now+1h`
5. **Escalate if needed:** `escalations create --incident <id> --path <path-id>`
6. **Communicate:** `status-pages incidents create --page <id> --name "..."`

### Always read before acting

- Check incident details before editing: `incidents get <id>`
- Check who's already assigned before escalating: look at `incident_role_assignments` in the get response
- Check current severity levels: `severities list`

### Error handling

Errors are JSON to stderr with a classification:
- `fixable_by: agent` — bad ID or missing flag. Read the hint and retry.
- `fixable_by: human` — credentials or permissions. Tell the user.
- `fixable_by: retry` — rate limit or server error. Wait and retry once.

## Quick Reference

```bash
# What's happening right now?
agent-incident incidents list --status active
agent-incident alerts list --status firing

# Investigate a specific incident
agent-incident incidents get <id>
agent-incident incidents updates <id>

# Who's on-call?
agent-incident schedules list
agent-incident schedules entries <schedule-id> --from now --to now+1h

# Respond
agent-incident escalations create --incident <id> --path <path-id>
agent-incident incidents edit <id> --summary "Root cause identified: ..."

# Create an incident (use severities list to find valid IDs)
agent-incident severities list
agent-incident incidents create --name "API latency spike" --severity <severity-id>

# Override on-call coverage
agent-incident schedules override <schedule-id> --user <user-id> --from now --to now+4h

# After resolution
agent-incident follow-ups list --incident <id>
agent-incident actions list --incident <id>

# Communicate externally
agent-incident status-pages list
agent-incident status-pages incidents create --page <id> --name "Degraded API performance"
agent-incident status-pages incidents update <sp-inc-id> --status resolved
```

## Key Concepts

- **Time formats**: relative (`now-15m`, `now-1h`, `now+1h`), RFC3339, or unix epoch
- **Output**: NDJSON for lists (one object per line), JSON for single items. `--full` for complete API response. `--format json|yaml|jsonl` to override
- **Compact mode**: List commands omit large fields (description, custom fields, timestamps) by default. Use `--full` to include everything
- **Pagination**: `--limit N` controls page size, `--after <cursor>` for next page. Cursor is returned in `@pagination` NDJSON line

## Deeper Reference

Per-domain details (only load when the quick reference above isn't enough):

```bash
agent-incident llm-help               # full command overview, all domains
agent-incident incidents llm-help     # incident lifecycle, create/edit fields
agent-incident alerts llm-help        # alert statuses, create alert events
agent-incident schedules llm-help     # schedule entries, overrides
agent-incident escalations llm-help   # escalation paths, create escalations
agent-incident status-pages llm-help  # status page incident management
```

## Discovery Commands

```bash
# Reference data (what values are valid?)
agent-incident severities list        # valid severity levels
agent-incident statuses list          # valid incident statuses
agent-incident roles list             # incident roles (lead, comms, etc.)
agent-incident custom-fields list     # org-specific custom fields

# Service catalog
agent-incident catalog types list
agent-incident catalog entries list --type <type-id> --query "checkout"

# People
agent-incident users list --query "alice"
agent-incident escalations paths list
```

## Auth Setup

If credentials aren't configured yet:
```bash
agent-incident auth add <alias> --api-key <key>
agent-incident auth check
```
Keys are managed at `https://app.incident.io/~/settings/api-keys`.

Environment variable also works: `INCIDENT_API_KEY=<key>`.

Multiple orgs: `agent-incident auth add prod --api-key <key>`, then `--organization prod` on any command.
