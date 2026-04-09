package roles

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/cli/shared"
)

func Register(root *cobra.Command, globals shared.GlobalsFunc) {
	role := &cobra.Command{
		Use:   "role",
		Short: "List and inspect incident roles",
	}

	registerList(role, globals)
	registerGet(role, globals)
	shared.RegisterLLMHelp(role, "LLM reference for incident roles", llmHelpText)

	root.AddCommand(role)
}

func registerList(parent *cobra.Command, globals shared.GlobalsFunc) {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all incident roles",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.APIKey, g.Org, g.Timeout, func(ctx context.Context, client *api.Client) error {
				items, err := client.ListIncidentRoles(ctx)
				if err != nil {
					return err
				}
				shared.WritePaginatedList(shared.ToAnySlice(items), nil, g.Format)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
}

func registerGet(parent *cobra.Command, globals shared.GlobalsFunc) {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get an incident role by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.APIKey, g.Org, g.Timeout, func(ctx context.Context, client *api.Client) error {
				item, err := client.GetIncidentRole(ctx, args[0])
				if err != nil {
					return err
				}
				shared.WriteItem(item, g.Format)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
}
