package customfields

import (
	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/cli/shared"
)

func registerLLMHelp(parent *cobra.Command) {
	shared.RegisterLLMHelp(parent, "Custom fields reference for LLMs", llmHelpText)
}

const llmHelpText = `agent-incident custom-fields — Custom field reference

COMMANDS
  custom-fields list
  custom-fields get <id>

NOTES
  Custom fields are organization-defined metadata attached to incidents.
  Read-only — field configuration is managed in the incident.io UI.
`
