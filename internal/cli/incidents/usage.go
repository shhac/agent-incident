package incidents

const incidentsLLMHelp = `# incidents domain — agent-incident CLI

## Commands

### incidents list
List incidents with optional filters.
  --status   Filter by status category: active, closed, paused, post_incident, declined, canceled
  --severity Filter by severity name (can specify multiple)
  --since    Show incidents created after this time (supports relative: now-1h, RFC3339, unix epoch)
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

### incidents edit <id>
Edit an existing incident.
  --status    New incident status
  --severity  New severity ID
  --summary   Updated summary

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
- Update status: incidents edit <id> --summary "Update: ..."
- Review timeline: incidents updates <id>
`
