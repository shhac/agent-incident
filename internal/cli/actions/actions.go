package actions

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/cli/shared"
)

func Register(root *cobra.Command, globals shared.GlobalsFunc) {
	cmd := &cobra.Command{
		Use:   "action",
		Short: "List and inspect incident actions",
	}

	registerList(cmd, globals)
	registerGet(cmd, globals)
	shared.RegisterLLMHelp(cmd, "Actions reference for LLMs", llmHelpText)

	root.AddCommand(cmd)
}

func registerList(parent *cobra.Command, globals shared.GlobalsFunc) {
	var incidentID, after string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List actions",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g, func(ctx context.Context, client *api.Client) error {
				items, cursor, err := client.ListActions(ctx, incidentID, limit, after)
				if err != nil {
					return err
				}
				shared.WritePaginatedList(shared.ToAnySlice(items), shared.CursorPagination(cursor), g.Format)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&incidentID, "incident", "", "Filter by incident ID")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum results")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor")
	parent.AddCommand(cmd)
}

func registerGet(parent *cobra.Command, globals shared.GlobalsFunc) {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get action details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g, func(ctx context.Context, client *api.Client) error {
				item, err := client.GetAction(ctx, args[0])
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
