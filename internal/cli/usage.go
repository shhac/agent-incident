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
  agent-incident incident list [--status <status>] [--severity <sev>] [--since <time>] [--full]
  agent-incident incident get <id-or-reference>       # accepts INC-2000, 2000, or UUID
  agent-incident incident create --name <text> --severity <sev-id> [--summary <text>]
  agent-incident incident edit <id> [--name <text>] [--severity <name-or-id>] [--status <name-or-id>] [--summary <text>] [--field "Name=value"] [--timestamp "Name=time"]
  agent-incident incident updates <id>

ALERTS
  agent-incident alert list [--status firing|resolved] [--source <src>] [--full]
  agent-incident alert get <id>
  agent-incident alert create --source-id <id> --title <text> [--description <text>]
  agent-incident alert incidents

ACTIONS & FOLLOW-UPS
  agent-incident action list [--incident <id>]
  agent-incident action get <id>
  agent-incident follow-up list [--incident <id>]
  agent-incident follow-up get <id>

ON-CALL & ESCALATION
  agent-incident oncall schedule list
  agent-incident oncall schedule get <id>
  agent-incident oncall schedule entries <id> [--from <time>] [--to <time>]
  agent-incident oncall schedule override <id> --user <id> --from <time> --to <time>
  agent-incident oncall escalation list
  agent-incident oncall escalation get <id>
  agent-incident oncall escalation create --incident <id> --path <id>
  agent-incident oncall escalation path list
  agent-incident oncall escalation path get <id>

STATUS PAGES
  agent-incident status-page list
  agent-incident status-page update list [--page <id>]
  agent-incident status-page update create --page <id> --name <text>
  agent-incident status-page update update <id> [--status <status>]

REFERENCE DATA
  agent-incident ref severity list / get <id>
  agent-incident ref status list / get <id>
  agent-incident ref role list / get <id>
  agent-incident ref user list [--query <text>] [--full] / get <id>
  agent-incident ref custom-field list / get <id>
  agent-incident ref timestamp list / get <id>
  agent-incident ref catalog types list / get <id>
  agent-incident ref catalog entries list [--type <id>] [--query <text>] / get <id>

TIME FORMATS
  Relative: now-15m, now-1h, now-1d, now+1h, now+30m
  Absolute: 2024-01-15T10:00:00Z (RFC3339)
  Unix epoch seconds

GLOBAL FLAGS
  --organization <alias>   Organization alias (or INCIDENT_API_KEY env)
  --api-key <key>          Direct API key override
  --format json|yaml|jsonl (default: jsonl for lists, json for single items)
  --timeout <ms>

Per-domain details: agent-incident <command> llm-help
`
