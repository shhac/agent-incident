package alerts

const alertsLLMHelp = `# alerts — Manage incident.io alerts

## Commands

### alerts list
List alerts with optional filters.
Flags:
  --status <s>    Filter by status (comma-separated: firing,resolved)
  --source <key>  Filter by deduplication key
  --from <time>   Filter by creation time (relative: now-1h, RFC3339, or epoch)
  --limit <n>     Page size (default 25)
  --after <cursor> Pagination cursor from previous response
  --full          Show full alert details (default: compact view)

Examples:
  agent-incident alerts list --status firing
  agent-incident alerts list --status firing,resolved --limit 10
  agent-incident alerts list --full --after <cursor>

### alerts get <id>
Retrieve a single alert by ID.

Examples:
  agent-incident alerts get 01HXYZ123

### alerts create
Create an alert event via an HTTP alert source.
Flags:
  --source-id <id>     Alert source config ID (required)
  --title <text>       Alert title (required)
  --description <text> Alert description

Examples:
  agent-incident alerts create --source-id 01HABC --title "CPU High" --description "CPU > 90%"

### alerts incidents
List alerts that are attached to incidents.
Flags:
  --limit <n>     Page size (default 25)
  --after <cursor> Pagination cursor

Examples:
  agent-incident alerts incidents --limit 50

## Output
- List commands output NDJSON (one object per line) by default.
- Get/create commands output JSON by default.
- Use --format yaml or --format jsonl to change.
- Compact view (default for list) shows: id, title, status, created_at.
- Use --full to see all fields including description, deduplication_key, source_url, updated_at.

## Pagination
When more results exist, a pagination line is emitted:
  {"@pagination": {"has_more": true, "next_cursor": "..."}}
Pass the cursor value to --after on the next call.
`
