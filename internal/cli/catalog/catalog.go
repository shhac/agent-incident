package catalog

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/cli/shared"
)

func Register(root *cobra.Command, globals shared.GlobalsFunc) {
	cmd := &cobra.Command{
		Use:   "catalog",
		Short: "Browse the service catalog",
	}

	types := &cobra.Command{
		Use:   "types",
		Short: "Manage catalog types",
	}
	registerTypesList(types, globals)
	registerTypesGet(types, globals)
	cmd.AddCommand(types)

	entries := &cobra.Command{
		Use:   "entries",
		Short: "Manage catalog entries",
	}
	registerEntriesList(entries, globals)
	registerEntriesGet(entries, globals)
	cmd.AddCommand(entries)

	shared.RegisterUsage(cmd, "catalog", usageText)
	root.AddCommand(cmd)
}

func registerTypesList(parent *cobra.Command, globals shared.GlobalsFunc) {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List catalog types",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g, func(ctx context.Context, client *api.Client) error {
				items, err := client.ListCatalogTypes(ctx)
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

func registerTypesGet(parent *cobra.Command, globals shared.GlobalsFunc) {
	cmd := &cobra.Command{
		Use:   "get <id>...",
		Short: "Get catalog type details (one or more IDs)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return shared.GetEntities(globals(), args, func(ctx context.Context, client *api.Client, id string) (any, error) {
				return client.GetCatalogType(ctx, id)
			})
		},
	}
	parent.AddCommand(cmd)
}

func registerEntriesList(parent *cobra.Command, globals shared.GlobalsFunc) {
	var typeID, query, after string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List catalog entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g, func(ctx context.Context, client *api.Client) error {
				items, cursor, err := client.ListCatalogEntries(ctx, typeID, query, limit, after)
				if err != nil {
					return err
				}
				shared.WritePaginatedList(shared.ToAnySlice(items), shared.CursorPagination(cursor), g.Format)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&typeID, "type", "", "Filter by catalog type ID")
	cmd.Flags().StringVar(&query, "query", "", "Search by name")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum results")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor")
	parent.AddCommand(cmd)
}

func registerEntriesGet(parent *cobra.Command, globals shared.GlobalsFunc) {
	cmd := &cobra.Command{
		Use:   "get <id>...",
		Short: "Get catalog entry details (one or more IDs)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return shared.GetEntities(globals(), args, func(ctx context.Context, client *api.Client, id string) (any, error) {
				return client.GetCatalogEntry(ctx, id)
			})
		},
	}
	parent.AddCommand(cmd)
}
