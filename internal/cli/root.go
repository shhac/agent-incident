package cli

import (
	"fmt"
	"os"

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
	"github.com/shhac/agent-incident/internal/cli/users"
)

var (
	flagOrg     string
	flagAPIKey  string
	flagFormat  string
	flagTimeout int
	flagDebug   bool
)

func allGlobals() *shared.GlobalFlags {
	return &shared.GlobalFlags{
		Org:     flagOrg,
		APIKey:  flagAPIKey,
		Format:  flagFormat,
		Timeout: flagTimeout,
		Debug:   flagDebug,
	}
}

func newRootCmd(version string) *cobra.Command {
	root := &cobra.Command{
		Use:           "agent-incident",
		Short:         "incident.io triage CLI for AI agents",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVar(&flagOrg, "organization", "", "Organization alias (or set INCIDENT_API_KEY)")
	root.PersistentFlags().StringVar(&flagAPIKey, "api-key", "", "API key (overrides stored credentials)")
	root.PersistentFlags().StringVar(&flagFormat, "format", "", "Output format: json, yaml, jsonl")
	root.PersistentFlags().IntVar(&flagTimeout, "timeout", 0, "Request timeout in milliseconds")
	root.PersistentFlags().BoolVar(&flagDebug, "debug", false, "Log HTTP requests and responses to stderr")

	registerLLMHelpCommand(root)
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
		Short: "Reference data lookups (severities, statuses, roles, users, fields, catalog)",
	}
	severities.Register(ref, allGlobals)
	statuses.Register(ref, allGlobals)
	roles.Register(ref, allGlobals)
	users.Register(ref, allGlobals)
	customfields.Register(ref, allGlobals)
	catalog.Register(ref, allGlobals)
	root.AddCommand(ref)

	return root
}

func Execute(version string) error {
	err := newRootCmd(version).Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return err
}
