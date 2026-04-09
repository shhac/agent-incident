package actions

const llmHelpText = `agent-incident actions — In-incident tasks

COMMANDS
  actions list [--incident <id>] [--limit N] [--after <cursor>]
  actions get <id>

NOTES
  Actions are tasks created during an incident (e.g., "restart service X").
  Filter by --incident to see actions for a specific incident.
`
