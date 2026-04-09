package statuses

const llmHelpText = `agent-incident statuses — Incident status lifecycle states

COMMANDS
  agent-incident statuses list          List all configured incident statuses
  agent-incident statuses get <id>      Get a single incident status by ID

FIELDS
  id            Unique status identifier
  name          Display name (e.g. "Investigating", "Resolved", "Closed")
  description   What this status means
  category      Status category: triage, active, post-incident, closed
  rank          Numeric rank for ordering within the lifecycle
  created_at    ISO 8601 creation timestamp
  updated_at    ISO 8601 last-update timestamp

NOTES
  Statuses are read-only configuration — they are defined in the incident.io
  dashboard and represent the lifecycle stages of an incident. Use
  "agent-incident incidents list --status <name>" to filter incidents by status,
  or "agent-incident incidents edit <id> --status <name>" to transition an
  incident to a different status.

  The category field groups statuses into lifecycle phases:
    triage         Initial assessment
    active         Actively being worked on
    post-incident  Resolved but follow-up pending
    closed         Fully closed

EXAMPLES
  # List all statuses to see the incident lifecycle
  agent-incident statuses list

  # Get details on a specific status
  agent-incident statuses get ST456
`
