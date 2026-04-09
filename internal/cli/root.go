package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/cli/auth"
	"github.com/shhac/agent-incident/internal/cli/shared"
)

var (
	flagOrg     string
	flagAPIKey  string
	flagFormat  string
	flagTimeout int
)

func allGlobals() *shared.GlobalFlags {
	return &shared.GlobalFlags{
		Org:     flagOrg,
		APIKey:  flagAPIKey,
		Format:  flagFormat,
		Timeout: flagTimeout,
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

	registerLLMHelpCommand(root)
	auth.Register(root)

	return root
}

func Execute(version string) error {
	err := newRootCmd(version).Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return err
}
