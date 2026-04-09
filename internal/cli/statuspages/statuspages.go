package statuspages

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/cli/shared"
)

func Register(root *cobra.Command, globals shared.GlobalsFunc) {
	cmd := &cobra.Command{
		Use:     "status-page",
		Aliases: []string{"statuspage"},
		Short:   "Manage status pages and their updates",
	}

	registerList(cmd, globals)

	update := &cobra.Command{
		Use:   "update",
		Short: "Manage status page updates",
	}
	registerIncidentsList(update, globals)
	registerIncidentsCreate(update, globals)
	registerIncidentsUpdate(update, globals)
	cmd.AddCommand(update)

	shared.RegisterLLMHelp(cmd, "Status pages reference for LLMs", llmHelpText)
	root.AddCommand(cmd)
}

func registerList(parent *cobra.Command, globals shared.GlobalsFunc) {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List status pages",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g, func(ctx context.Context, client *api.Client) error {
				items, err := client.ListStatusPages(ctx)
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

func registerIncidentsList(parent *cobra.Command, globals shared.GlobalsFunc) {
	var pageRef string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List status page incidents",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g, func(ctx context.Context, client *api.Client) error {
				pageID := pageRef
				if pageRef != "" {
					resolved, err := shared.ResolveStatusPageID(ctx, client, pageRef)
					if err != nil {
						return err
					}
					pageID = resolved
				}
				items, err := client.ListStatusPageIncidents(ctx, pageID)
				if err != nil {
					return err
				}
				shared.WritePaginatedList(shared.ToAnySlice(items), nil, g.Format)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&pageRef, "page", "", "Filter by status page name or ID")
	parent.AddCommand(cmd)
}

func registerIncidentsCreate(parent *cobra.Command, globals shared.GlobalsFunc) {
	var pageRef, name string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a status page incident",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			pageOk := shared.RequireFlag("page", pageRef, "")
			nameOk := shared.RequireFlag("name", name, "")
			if !pageOk || !nameOk {
				return nil
			}
			return shared.WithClient(g, func(ctx context.Context, client *api.Client) error {
				pageID, err := shared.ResolveStatusPageID(ctx, client, pageRef)
				if err != nil {
					return err
				}
				item, err := client.CreateStatusPageIncident(ctx, api.CreateStatusPageIncidentParams{
					StatusPageID: pageID,
					Name:         name,
				})
				if err != nil {
					return err
				}
				shared.WriteItem(item, g.Format)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&pageRef, "page", "", "Status page name or ID (required)")
	cmd.Flags().StringVar(&name, "name", "", "Incident name (required)")
	parent.AddCommand(cmd)
}

func registerIncidentsUpdate(parent *cobra.Command, globals shared.GlobalsFunc) {
	var status string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a status page incident",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g, func(ctx context.Context, client *api.Client) error {
				item, err := client.UpdateStatusPageIncident(ctx, args[0], api.UpdateStatusPageIncidentParams{
					Status: status,
				})
				if err != nil {
					return err
				}
				shared.WriteItem(item, g.Format)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&status, "status", "", "New status")
	parent.AddCommand(cmd)
}
