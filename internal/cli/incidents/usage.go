package incidents

const incidentsLLMHelp = `# incidents domain — agent-incident CLI

## Commands

### incidents list
List incidents with optional filters.
  --status   Filter by status category: active, closed, paused, post_incident, declined, canceled
  --severity Filter by severity name (can specify multiple)
  --from     Show incidents created after this time (supports relative: now-1h, RFC3339, unix epoch)
  --limit    Page size (default 25)
  --after    Pagination cursor from previous response
  --full     Return full incident objects (default: compact with id, name, status, severity, created_at, incident_lead)

### incidents get <id>
Retrieve a single incident with all fields including role assignments, custom fields, timestamps, and external resources.

### incidents create
Create a new incident.
  --name      Incident name (required)
  --severity  Severity ID (use severity list to find IDs)
  --summary   Incident summary text

### incidents edit <id-or-reference>
Edit an existing incident (accepts INC-2000, 2000, or UUID).
  --name       New incident name
  --status     New incident status (name or ID, e.g. "Closed", "Investigating")
  --severity   Severity (name or ID, e.g. "Critical", "SEV1")
  --summary    Updated summary
  --field      Set custom field value (repeatable): "Field Name=value"
               For select fields, resolves option names. Empty value clears the field.
  --timestamp  Set timestamp value (repeatable): "Reported at=2026-04-09T15:00:00Z"
               Supports relative times (now, now-1h), RFC3339, unix epoch. Empty value clears.

### incidents updates <id>
List status updates posted to an incident.
  --limit  Page size (default 25)
  --after  Pagination cursor

## Pagination
List commands return NDJSON by default. When more results exist, a final line contains:
  {"@pagination": {"has_more": true, "next_cursor": "..."}}
Pass the cursor value to --after for the next page.

## Compact vs Full
By default, "incidents list" returns compact objects (id, name, status, severity, created_at, incident_lead).
Use --full to get complete incident objects with all nested data.

## Common Workflows
- List active incidents: incidents list --status active
- Get incident details: incidents get <id>
- Create and triage: incidents create --name "..." --severity <sev_id>
- Update status: incidents edit <id> --status Closed --summary "Resolved: ..."
- Set severity by name: incidents edit <id> --severity Critical
- Set custom fields: incidents edit <id> --field "Affected Team=Platform" --field "Root Cause=DNS"
- Set timestamps: incidents edit <id> --timestamp "Resolved at=2026-04-09T15:30:00Z"
- Review timeline: incidents updates <id>

## Discovering Valid Values
- Severities: ref severity list
- Statuses: ref status list
- Custom fields and options: ref custom-field list
- Timestamp definitions: ref timestamp list
`
