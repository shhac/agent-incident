package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func registerUsageCommand(root *cobra.Command) {
	root.AddCommand(&cobra.Command{
		Use:   "usage",
		Short: "LLM-optimized reference card",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(usageText)
		},
	})
}

const usageText = `agent-incident — incident.io triage CLI for AI agents

AUTH SETUP
  agent-incident auth add <alias> --form    # preferred: prompt for the key via a native OS dialog (key never seen by the agent)
  agent-incident auth add <alias>           # non-interactive: pipe the key on stdin (keeps it off argv)
  # Avoid --api-key <key> with a pasted secret — it lands on argv, in shell history, and in the agent transcript.
  agent-incident auth check [alias]
  agent-incident auth default <alias>
  agent-incident auth list
  agent-incident auth remove <alias>

INCIDENTS (triage starting point)
  agent-incident incident list [--status <status>] [--severity <sev>] [--from <time>] [--full]
  agent-incident incident get <id-or-reference>...    # accepts INC-2000, 2000, or UUID; 1..N ids
  agent-incident incident create --name <text> --severity <sev-id> [--summary <text>]
  agent-incident incident edit <id> [--name <text>] [--severity <name-or-id>] [--status <name-or-id>] [--summary <text>] [--field "Name=value"] [--timestamp "Name=time"]
  agent-incident incident updates <id>

ALERTS
  agent-incident alert list [--status firing|resolved] [--source <src>] [--full]
  agent-incident alert get <id>...
  agent-incident alert create --source-id <id> --title <text> [--description <text>]
  agent-incident alert incidents

ACTIONS & FOLLOW-UPS
  agent-incident action list [--incident <id>]
  agent-incident action get <id>...
  agent-incident follow-up list [--incident <id>]
  agent-incident follow-up get <id>...

ON-CALL & ESCALATION
  agent-incident oncall schedule list
  agent-incident oncall schedule get <id>...
  agent-incident oncall schedule entries <id> [--from <time>] [--to <time>]
  agent-incident oncall schedule override <id> --user <id> --from <time> --to <time>
  agent-incident oncall escalation list
  agent-incident oncall escalation get <id>...
  agent-incident oncall escalation create --incident <id> --path <id>
  agent-incident oncall escalation path list
  agent-incident oncall escalation path get <id>...

STATUS PAGES
  agent-incident status-page list
  agent-incident status-page update list [--page <id>]
  agent-incident status-page update create --page <id> --name <text>
  agent-incident status-page update update <id> [--status <status>]

REFERENCE DATA
  agent-incident ref severity list / get <id>...
  agent-incident ref status list / get <id>...
  agent-incident ref role list / get <id>...
  agent-incident ref user list [--query <text>] [--full] / get <id>...
  agent-incident ref custom-field list / get <id>...
  agent-incident ref timestamp list / get <id>...
  agent-incident ref catalog types list / get <id>...
  agent-incident ref catalog entries list [--type <id>] [--query <text>] / get <id>...

TIME FORMATS
  Relative: now-15m, now-1h, now-1d, now+1h, now+30m
  Absolute: 2024-01-15T10:00:00Z (RFC3339)
  Unix epoch seconds

GLOBAL FLAGS
  --org/-o <alias>         Organization alias (or INCIDENT_API_KEY env)
  --api-key <key>          Direct API key override
  --format json|yaml|jsonl (default: jsonl for lists and gets; --format json for a pretty object)
  --timeout <ms>
  -d/--debug               Log HTTP diagnostics to stderr

OUTPUT
  Get (single + multi). get <id>... takes one or more ids and returns one result per id, in
  input order. Default output is NDJSON: one line per id — the record, or
  {"@unresolved":{"id","reason","fixable_by","hint"?}} for an id that couldn't be resolved
  (e.g. not found / bad id). --format json|yaml collapses to one {"data":[...], "@unresolved":[...]}
  envelope. A single get <id> is just the one-element case (NDJSON one line by default; pass
  --format json for the object). Item-level misses stay on stdout and exit 0; only a
  command-level failure (auth, network) goes to stderr with exit 1 and empty stdout.

Per-domain details: agent-incident <command> usage
`
