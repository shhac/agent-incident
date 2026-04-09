package actions

import (
	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/cli/shared"
)

func registerLLMHelp(parent *cobra.Command) {
	shared.RegisterLLMHelp(parent, "Actions reference for LLMs", llmHelpText)
}

const llmHelpText = `agent-incident actions — In-incident tasks

COMMANDS
  actions list [--incident <id>] [--limit N] [--after <cursor>]
  actions get <id>

NOTES
  Actions are tasks created during an incident (e.g., "restart service X").
  Filter by --incident to see actions for a specific incident.
`
