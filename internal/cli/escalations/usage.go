package escalations

import (
	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/cli/shared"
)

const escalationsLLMHelp = `# escalations domain — agent-incident CLI

## Commands

### escalations list
List all escalations.

### escalations get <id>
Retrieve a single escalation by ID, including title, status, and timestamps.

### escalations create
Create a new escalation for an incident.
  --incident  Incident ID (required)
  --path      Escalation path ID (required)

### escalations paths list
List all configured escalation paths.

### escalations paths get <id>
Retrieve a single escalation path by ID, including the path configuration.

## Common Workflows
- Page someone for an incident: escalations create --incident <id> --path <path-id>
- See available paths: escalations paths list
- Review escalation details: escalations get <id>
- List all escalations: escalations list
`

func registerLLMHelp(parent *cobra.Command) {
	shared.RegisterLLMHelp(parent, "LLM reference for escalations commands", escalationsLLMHelp)
}
