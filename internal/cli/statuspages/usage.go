package statuspages

import (
	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/cli/shared"
)

func registerLLMHelp(parent *cobra.Command) {
	shared.RegisterLLMHelp(parent, "Status pages reference for LLMs", llmHelpText)
}

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
