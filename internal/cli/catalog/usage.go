package catalog

const llmHelpText = `agent-incident catalog — Service catalog queries

COMMANDS
  catalog types list
  catalog types get <id>
  catalog entries list [--type <id>] [--query <text>] [--limit N] [--after <cursor>]
  catalog entries get <id>

NOTES
  The catalog maps services, teams, and other entities to incidents.
  Use --type to filter entries by catalog type.
  Use --query to search entries by name.
`
