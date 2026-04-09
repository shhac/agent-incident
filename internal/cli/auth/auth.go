package auth

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/cli/shared"
	"github.com/shhac/agent-incident/internal/config"
	"github.com/shhac/agent-incident/internal/credential"
	"github.com/shhac/agent-incident/internal/output"
)

// Register adds the auth command group to the root command.
func Register(root *cobra.Command) {
	auth := &cobra.Command{
		Use:   "auth",
		Short: "Manage API credentials and organizations",
	}

	registerAdd(auth)
	registerCheck(auth)
	registerDefault(auth)
	registerList(auth)
	registerRemove(auth)

	root.AddCommand(auth)
}

func registerAdd(parent *cobra.Command) {
	var apiKey string

	cmd := &cobra.Command{
		Use:   "add <alias>",
		Short: "Add an organization with an API key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]
			if !shared.RequireFlag("api-key", apiKey, "Provide --api-key <key>") {
				return nil
			}

			storage, err := credential.Store(alias, credential.Credential{APIKey: apiKey})
			if err != nil {
				output.WriteError(os.Stderr, err)
				return nil
			}

			if err := config.StoreOrganization(alias); err != nil {
				output.WriteError(os.Stderr, err)
				return nil
			}

			shared.WriteItem(map[string]any{
				"status":  "added",
				"alias":   alias,
				"storage": storage,
			}, "")
			return nil
		},
	}
	cmd.Flags().StringVar(&apiKey, "api-key", "", "API key for this organization (required)")
	parent.AddCommand(cmd)
}

func registerCheck(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "check [alias]",
		Short: "Verify stored credentials by calling the identity endpoint",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var alias string
			if len(args) > 0 {
				alias = args[0]
			}

			// Resolve which credential to check
			var apiKey string
			if alias != "" {
				cred, err := credential.Get(alias)
				if err != nil {
					output.WriteError(os.Stderr, err)
					return nil
				}
				apiKey = cred.APIKey
			} else {
				// Try env first, then default org
				apiKey = os.Getenv("INCIDENT_API_KEY")
				if apiKey == "" {
					cfg := config.Read()
					if cfg.DefaultOrg == "" {
						output.WriteError(os.Stderr, fmt.Errorf("no alias specified and no default organization configured"))
						return nil
					}
					alias = cfg.DefaultOrg
					cred, err := credential.Get(alias)
					if err != nil {
						output.WriteError(os.Stderr, err)
						return nil
					}
					apiKey = cred.APIKey
				}
			}

			var client *api.Client
			if apiURL := os.Getenv("INCIDENT_API_URL"); apiURL != "" {
				client = api.NewTestClient(apiURL, apiKey)
			} else {
				client = api.NewClient(apiKey)
			}

			// Use ClientFactory for test injection
			if shared.ClientFactory != nil {
				var err error
				client, err = shared.ClientFactory()
				if err != nil {
					output.WriteError(os.Stderr, err)
					return nil
				}
			}

			ctx, cancel := shared.MakeContext(0)
			defer cancel()

			identity, err := client.GetIdentity(ctx)
			if err != nil {
				output.WriteError(os.Stderr, err)
				return nil
			}

			result := map[string]any{
				"status":   "ok",
				"identity": identity,
			}
			if alias != "" {
				result["alias"] = alias
			}
			shared.WriteItem(result, "")
			return nil
		},
	}
	parent.AddCommand(cmd)
}

func registerDefault(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "default <alias>",
		Short: "Set the default organization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]
			if err := config.SetDefault(alias); err != nil {
				output.WriteError(os.Stderr, err)
				return nil
			}
			shared.WriteItem(map[string]any{
				"status":  "default_set",
				"alias":   alias,
			}, "")
			return nil
		},
	}
	parent.AddCommand(cmd)
}

func registerList(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configured organizations",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Read()
			orgs := make([]map[string]any, 0)
			for alias := range cfg.Organizations {
				org := map[string]any{
					"alias":   alias,
					"default": alias == cfg.DefaultOrg,
				}
				orgs = append(orgs, org)
			}
			shared.WritePaginatedList(shared.ToAnySlice(orgs), nil, "")
			return nil
		},
	}
	parent.AddCommand(cmd)
}

func registerRemove(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "remove <alias>",
		Short: "Remove an organization and its credentials",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]
			if err := credential.Remove(alias); err != nil {
				output.WriteError(os.Stderr, err)
				return nil
			}
			if err := config.RemoveOrganization(alias); err != nil {
				output.WriteError(os.Stderr, err)
				return nil
			}
			shared.WriteItem(map[string]any{
				"status": "removed",
				"alias":  alias,
			}, "")
			return nil
		},
	}
	parent.AddCommand(cmd)
}
