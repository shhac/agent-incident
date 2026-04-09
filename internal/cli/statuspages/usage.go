package statuspages

const llmHelpText = `agent-incident status-pages — External status page management

COMMANDS
  status-pages list
  status-pages incidents list [--page <id>]
  status-pages incidents create --page <id> --name <text>
  status-pages incidents update <id> [--status <status>]

NOTES
  Status pages communicate incident impact to external stakeholders.
  Use --page to filter incidents by a specific status page.
`
