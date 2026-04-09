package schedules

import (
	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/cli/shared"
)

const schedulesLLMHelp = `# schedules domain — agent-incident CLI

## Commands

### schedules list
List all on-call schedules.

### schedules get <id>
Retrieve a single schedule by ID, including timezone and metadata.

### schedules entries <schedule-id>
List schedule entries (who is on-call) for a time window.
  --from  Window start (supports relative: now-1h, RFC3339, unix epoch; default: now-1h)
  --to    Window end (supports relative: now+1h, RFC3339, unix epoch; default: now)

### schedules override <schedule-id>
Create an override on a schedule (temporarily replace who is on-call).
  --user  User ID for the override (required)
  --from  Override start time (required; supports relative, RFC3339, epoch)
  --to    Override end time (required; supports relative, RFC3339, epoch)

## Common Workflows
- See who is on-call now: schedules entries <schedule-id>
- See on-call for next 24h: schedules entries <id> --from now --to now+24h
- Cover for someone: schedules override <id> --user <user-id> --from now --to now+4h
- List all schedules: schedules list
`

func registerLLMHelp(parent *cobra.Command) {
	shared.RegisterLLMHelp(parent, "LLM reference for schedules commands", schedulesLLMHelp)
}
