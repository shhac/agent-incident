package timestamps

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/cli/shared"
)

func Register(root *cobra.Command, globals shared.GlobalsFunc) {
	ts := &cobra.Command{
		Use:   "timestamp",
		Short: "List and inspect incident timestamps",
	}

	registerList(ts, globals)
	registerGet(ts, globals)
	shared.RegisterUsage(ts, "timestamp", usageText)

	root.AddCommand(ts)
}

func registerList(parent *cobra.Command, globals shared.GlobalsFunc) {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all incident timestamp definitions",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g, func(ctx context.Context, client *api.Client) error {
				items, err := client.ListIncidentTimestamps(ctx)
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
		Use:   "get <id>...",
		Short: "Get a timestamp definition by ID (one or more IDs)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return shared.GetEntities(globals(), args, func(ctx context.Context, client *api.Client, id string) (any, error) {
				return client.GetIncidentTimestamp(ctx, id)
			})
		},
	}
	parent.AddCommand(cmd)
}
