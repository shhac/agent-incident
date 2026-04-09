package followups

import (
	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/cli/shared"
)

func registerLLMHelp(parent *cobra.Command) {
	shared.RegisterLLMHelp(parent, "Follow-ups reference for LLMs", llmHelpText)
}

const llmHelpText = `agent-incident follow-ups — Post-incident action items

COMMANDS
  follow-ups list [--incident <id>] [--limit N] [--after <cursor>]
  follow-ups get <id>

NOTES
  Follow-ups are action items created after an incident is resolved.
  Filter by --incident to see follow-ups for a specific incident.
`
