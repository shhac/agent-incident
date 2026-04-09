package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func registerLLMHelpCommand(root *cobra.Command) {
	root.AddCommand(&cobra.Command{
		Use:   "llm-help",
		Short: "LLM-optimized reference card",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(usageText)
		},
	})
}

const usageText = `agent-incident — incident.io triage CLI for AI agents

AUTH SETUP
  agent-incident auth add <alias> --api-key <key>
  agent-incident auth check [alias]
  agent-incident auth default <alias>
  agent-incident auth list
  agent-incident auth remove <alias>

INCIDENTS (triage starting point)
  agent-incident incidents list [--status <status>] [--severity <sev>] [--since <time>] [--full]
  agent-incident incidents get <id>
  agent-incident incidents create --name <text> --severity <sev> [--summary <text>]
  agent-incident incidents edit <id> [--status <status>] [--severity <sev>] [--summary <text>]
  agent-incident incidents updates <id>

ALERTS
  agent-incident alerts list [--status firing|resolved] [--source <src>] [--since <time>] [--full]
  agent-incident alerts get <id>
  agent-incident alerts create --source-id <id> --title <text> [--description <text>]
  agent-incident alerts incidents

SEVERITIES
  agent-incident severities list
  agent-incident severities get <id>

STATUSES
  agent-incident statuses list
  agent-incident statuses get <id>

USERS
  agent-incident users list [--query <text>] [--full]
  agent-incident users get <id>

ROLES
  agent-incident roles list
  agent-incident roles get <id>

SCHEDULES
  agent-incident schedules list
  agent-incident schedules get <id>
  agent-incident schedules entries <id> [--from <time>] [--to <time>]
  agent-incident schedules override <id> --user <id> --from <time> --to <time>

ESCALATIONS
  agent-incident escalations list
  agent-incident escalations get <id>
  agent-incident escalations create --incident <id> --path <id>
  agent-incident escalations paths list
  agent-incident escalations paths get <id>

ACTIONS
  agent-incident actions list [--incident <id>]
  agent-incident actions get <id>

FOLLOW-UPS
  agent-incident follow-ups list [--incident <id>]
  agent-incident follow-ups get <id>

CATALOG
  agent-incident catalog types list
  agent-incident catalog types get <id>
  agent-incident catalog entries list --type <id> [--query <text>]
  agent-incident catalog entries get <id>

CUSTOM FIELDS
  agent-incident custom-fields list
  agent-incident custom-fields get <id>

STATUS PAGES
  agent-incident status-pages list
  agent-incident status-pages incidents list [--page <id>]
  agent-incident status-pages incidents create --page <id> --name <text>
  agent-incident status-pages incidents update <id> --status <status>

TIME FORMATS
  Relative: now-15m, now-1h, now-1d, now+1h, now+30m
  Absolute: 2024-01-15T10:00:00Z (RFC3339)
  Unix epoch seconds

GLOBAL FLAGS
  --organization <alias>   Organization alias (or INCIDENT_API_KEY env)
  --api-key <key>          Direct API key override
  --format json|yaml|jsonl (default: jsonl for lists, json for single items)
  --timeout <ms>

Per-domain details: agent-incident <domain> llm-help
`
