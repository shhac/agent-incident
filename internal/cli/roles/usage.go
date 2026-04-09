package roles

const llmHelpText = `agent-incident roles — Incident role definitions

COMMANDS
  agent-incident roles list          List all configured incident roles
  agent-incident roles get <id>      Get a single incident role by ID

FIELDS
  id             Unique role identifier
  name           Display name (e.g. "Incident Lead", "Communications Lead")
  description    What this role is responsible for
  required       Whether this role must be assigned during an incident
  instructions   Guidance shown to the person assigned this role
  shortform      Abbreviated label for the role
  created_at     ISO 8601 creation timestamp
  updated_at     ISO 8601 last-update timestamp

NOTES
  Incident roles are read-only configuration — they are defined in the
  incident.io dashboard and assigned to users during incidents. Use
  "agent-incident incidents get <id>" to see role assignments on an incident.

EXAMPLES
  # List all incident roles
  agent-incident roles list

  # Get details on a specific role
  agent-incident roles get ROLE123
`
