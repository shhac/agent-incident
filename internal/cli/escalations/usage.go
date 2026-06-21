package escalations

const usageText = `# escalations domain — agent-incident CLI

## Commands

### escalations list
List all escalations.

### escalations get <id>...
Retrieve one or more escalations by ID (1..N ids). Default output is NDJSON — one line per id:
the record, or {"@unresolved":{...}} for an id not found.

### escalations create
Create a new escalation for an incident.
  --incident  Incident ID (required)
  --path      Escalation path name or ID (required)

### escalations paths list
List all configured escalation paths.

### escalations paths get <id>...
Retrieve one or more escalation paths by ID (1..N ids). Default output is NDJSON — one line per
id: the record, or {"@unresolved":{...}} for an id not found.

## Name Resolution
The --path flag accepts a name or ID. If the value doesn't look like a ULID,
it is matched against escalation path names (case-insensitive, substring match).

## Common Workflows
- Page someone for an incident: escalations create --incident <id> --path "Primary Path"
- See available paths: escalations paths list
- Review escalation details: escalations get <id>...
- List all escalations: escalations list
`
