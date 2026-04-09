package users

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/cli/shared"
)

func Register(root *cobra.Command, globals shared.GlobalsFunc) {
	usr := &cobra.Command{
		Use:   "users",
		Short: "List and inspect incident.io users",
	}

	registerList(usr, globals)
	registerGet(usr, globals)
	shared.RegisterLLMHelp(usr, "LLM reference for users", llmHelpText)

	root.AddCommand(usr)
}

func registerList(parent *cobra.Command, globals shared.GlobalsFunc) {
	var query string
	var limit int
	var after string
	var full bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List users",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.APIKey, g.Org, g.Timeout, func(ctx context.Context, client *api.Client) error {
				users, pagMeta, err := client.ListUsers(ctx, query, limit, after)
				if err != nil {
					return err
				}
				if full {
					shared.WritePaginatedList(shared.ToAnySlice(users), shared.CursorPagination(pagMeta.After), g.Format)
				} else {
					shared.WritePaginatedList(shared.ToAnySlice(api.ToCompactUsers(users)), shared.CursorPagination(pagMeta.After), g.Format)
				}
				return nil
			})
		},
	}

	cmd.Flags().StringVar(&query, "query", "", "Filter users by name or email")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of users to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")
	cmd.Flags().BoolVar(&full, "full", false, "Show all fields (default: compact view)")

	parent.AddCommand(cmd)
}

func registerGet(parent *cobra.Command, globals shared.GlobalsFunc) {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a user by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.APIKey, g.Org, g.Timeout, func(ctx context.Context, client *api.Client) error {
				item, err := client.GetUser(ctx, args[0])
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
