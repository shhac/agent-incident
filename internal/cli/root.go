package cli

import (
	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/cli/actions"
	"github.com/shhac/agent-incident/internal/cli/alerts"
	"github.com/shhac/agent-incident/internal/cli/auth"
	"github.com/shhac/agent-incident/internal/cli/catalog"
	"github.com/shhac/agent-incident/internal/cli/customfields"
	"github.com/shhac/agent-incident/internal/cli/escalations"
	"github.com/shhac/agent-incident/internal/cli/followups"
	"github.com/shhac/agent-incident/internal/cli/incidents"
	"github.com/shhac/agent-incident/internal/cli/roles"
	"github.com/shhac/agent-incident/internal/cli/schedules"
	"github.com/shhac/agent-incident/internal/cli/severities"
	"github.com/shhac/agent-incident/internal/cli/shared"
	"github.com/shhac/agent-incident/internal/cli/statuses"
	"github.com/shhac/agent-incident/internal/cli/statuspages"
	"github.com/shhac/agent-incident/internal/cli/timestamps"
	"github.com/shhac/agent-incident/internal/cli/users"
	"github.com/shhac/agent-incident/internal/output"
	libcli "github.com/shhac/lib-agent-cli/cli"
	agentmcp "github.com/shhac/lib-agent-mcp"
)

func newRootCmd(version string) *cobra.Command {
	g := &shared.GlobalFlags{}
	allGlobals := func() *shared.GlobalFlags { return g }

	root := libcli.NewRoot(libcli.Options{
		Use:           "agent-incident",
		Short:         "incident.io triage CLI for AI agents",
		Version:       version,
		Globals:       &g.Globals,
		DefaultFormat: output.FormatNDJSON,
		UnknownHint:   "run 'agent-incident usage' to see the available commands",
	})

	pf := root.PersistentFlags()
	pf.StringVarP(&g.Org, "org", "o", "", "Organization alias (or set INCIDENT_API_KEY)")
	pf.StringVar(&g.APIKey, "api-key", "", "API key (overrides stored credentials)")

	registerUsageCommand(root)
	auth.Register(root)

	// Core triage
	incidents.Register(root, allGlobals)
	alerts.Register(root, allGlobals)
	actions.Register(root, allGlobals)
	followups.Register(root, allGlobals)

	// On-call & escalation
	oncall := &cobra.Command{
		Use:   "oncall",
		Short: "On-call schedules, overrides, and escalations",
	}
	schedules.Register(oncall, allGlobals)
	escalations.Register(oncall, allGlobals)
	root.AddCommand(oncall)

	// Status pages
	statuspages.Register(root, allGlobals)

	// Reference data
	ref := &cobra.Command{
		Use:   "ref",
		Short: "Reference data lookups (severities, statuses, roles, users, fields, timestamps, catalog)",
	}
	severities.Register(ref, allGlobals)
	statuses.Register(ref, allGlobals)
	roles.Register(ref, allGlobals)
	users.Register(ref, allGlobals)
	customfields.Register(ref, allGlobals)
	timestamps.Register(ref, allGlobals)
	catalog.Register(ref, allGlobals)
	root.AddCommand(ref)

	installGroupUnknownHandlers(root)

	// Expose the whole command tree as an MCP server (added last, so it reflects
	// the complete tree). --color/--expose are output-shaping, irrelevant to a
	// tool call, so hide them from the generated schemas.
	root.AddCommand(agentmcp.Command(root, agentmcp.WithHiddenFlags("color", "expose")))

	return root
}

// installGroupUnknownHandlers walks the assembled tree and installs the
// structured unknown-subcommand handler on every group command (one that owns
// subcommands but has no Run of its own), so `agent-incident <group> bogus`
// returns a fixable_by:agent error and exits non-zero instead of cobra's usage
// text. The root already has its handler from libcli.NewRoot, so it is skipped.
func installGroupUnknownHandlers(root *cobra.Command) {
	var walk func(cmd *cobra.Command)
	walk = func(cmd *cobra.Command) {
		for _, sub := range cmd.Commands() {
			if sub.HasSubCommands() && sub.Run == nil && sub.RunE == nil {
				libcli.HandleUnknownCommand(sub, "run 'agent-incident usage' to see the available commands")
			}
			walk(sub)
		}
	}
	walk(root)
}

// Run builds the root command and runs it, rendering any error as the family's
// structured stderr line and exiting non-zero on failure.
func Run(version string) {
	libcli.Run(newRootCmd(version))
}
