package customfields

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/cli/shared"
)

func Register(root *cobra.Command, globals shared.GlobalsFunc) {
	cmd := &cobra.Command{
		Use:     "custom-fields",
		Aliases: []string{"customfields"},
		Short:   "List and inspect custom fields",
	}

	registerList(cmd, globals)
	registerGet(cmd, globals)
	registerLLMHelp(cmd)

	root.AddCommand(cmd)
}

func registerList(parent *cobra.Command, globals shared.GlobalsFunc) {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List custom fields",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.APIKey, g.Org, g.Timeout, func(ctx context.Context, client *api.Client) error {
				items, err := client.ListCustomFields(ctx)
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
		Short: "Get custom field details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.APIKey, g.Org, g.Timeout, func(ctx context.Context, client *api.Client) error {
				item, err := client.GetCustomField(ctx, args[0])
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
