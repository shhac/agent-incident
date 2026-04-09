package severities

const llmHelpText = `agent-incident severities — Incident severity levels

COMMANDS
  agent-incident severities list          List all configured severities
  agent-incident severities get <id>      Get a single severity by ID

FIELDS
  id            Unique severity identifier
  name          Display name (e.g. "Critical", "Major", "Minor")
  description   What this severity level means
  rank          Numeric rank for ordering (lower = more severe)
  created_at    ISO 8601 creation timestamp
  updated_at    ISO 8601 last-update timestamp

NOTES
  Severities are read-only configuration — they are defined in the incident.io
  dashboard and referenced by incidents. Use "agent-incident incidents list
  --severity <name>" to filter incidents by severity.

EXAMPLES
  # List all severities to see available options
  agent-incident severities list

  # Get details on a specific severity
  agent-incident severities get SEV123
`
