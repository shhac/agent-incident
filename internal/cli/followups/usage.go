package followups

const llmHelpText = `agent-incident follow-ups — Post-incident action items

COMMANDS
  follow-ups list [--incident <id>] [--limit N] [--after <cursor>]
  follow-ups get <id>

NOTES
  Follow-ups are action items created after an incident is resolved.
  Filter by --incident to see follow-ups for a specific incident.
`
