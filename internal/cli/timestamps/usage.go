package timestamps

const llmHelpText = `agent-incident timestamps — Incident timestamp definitions

COMMANDS
  agent-incident timestamp list          List all timestamp definitions
  agent-incident timestamp get <id>      Get a single timestamp definition by ID

FIELDS
  id    Unique timestamp identifier
  name  Display name (e.g. "Reported at", "Resolved at", "Mitigated at")
  rank  Numeric rank for ordering

NOTES
  Timestamp definitions describe the timestamps an organization tracks on
  incidents (e.g. "Reported at", "Resolved at"). These are read-only
  configuration defined in the incident.io dashboard.

  To set timestamp values on an incident, use:
    agent-incident incident edit <id> --timestamp "Resolved at=2026-04-09T15:30:00Z"

EXAMPLES
  # List all timestamp definitions to see available names
  agent-incident timestamp list

  # Get details on a specific timestamp
  agent-incident timestamp get 01TS00RESOLVEDAAAAAAAAAAAAA
`
