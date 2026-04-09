package followups

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/cli/shared"
)

func Register(root *cobra.Command, globals shared.GlobalsFunc) {
	cmd := &cobra.Command{
		Use:     "follow-up",
		Aliases: []string{"followup"},
		Short:   "List and inspect post-incident follow-ups",
	}

	registerList(cmd, globals)
	registerGet(cmd, globals)
	registerLLMHelp(cmd)

	root.AddCommand(cmd)
}

func registerList(parent *cobra.Command, globals shared.GlobalsFunc) {
	var incidentID, after string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List follow-ups",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g, func(ctx context.Context, client *api.Client) error {
				items, cursor, err := client.ListFollowUps(ctx, incidentID, limit, after)
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
		Short: "Get follow-up details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g, func(ctx context.Context, client *api.Client) error {
				item, err := client.GetFollowUp(ctx, args[0])
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
