package statuspages

const llmHelpText = `agent-incident status-pages — External status page management

COMMANDS
  status-pages list
  status-pages incidents list [--page <name-or-id>]
  status-pages incidents create --page <name-or-id> --name <text>
  status-pages incidents update <id> [--status <status>]

NOTES
  Status pages communicate incident impact to external stakeholders.
  The --page flag accepts a name or ID. If the value doesn't look like a ULID,
  it is matched against status page names (case-insensitive, substring match).
`
