package escalations

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/cli/shared"
)

// Register adds the escalations command group to the root command.
func Register(root *cobra.Command, globals shared.GlobalsFunc) {
	escalations := &cobra.Command{
		Use:   "escalation",
		Short: "Manage escalations and escalation paths",
	}

	registerList(escalations, globals)
	registerGet(escalations, globals)
	registerCreate(escalations, globals)
	registerPaths(escalations, globals)
	registerLLMHelp(escalations)

	root.AddCommand(escalations)
}

func registerList(parent *cobra.Command, globals shared.GlobalsFunc) {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List escalations",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.APIKey, g.Org, g.Timeout, func(ctx context.Context, client *api.Client) error {
				escalations, err := client.ListEscalations(ctx)
				if err != nil {
					return err
				}
				shared.WritePaginatedList(shared.ToAnySlice(escalations), nil, g.Format)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
}

func registerGet(parent *cobra.Command, globals shared.GlobalsFunc) {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a single escalation by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.APIKey, g.Org, g.Timeout, func(ctx context.Context, client *api.Client) error {
				escalation, err := client.GetEscalation(ctx, args[0])
				if err != nil {
					return err
				}
				shared.WriteItem(escalation, g.Format)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
}

func registerCreate(parent *cobra.Command, globals shared.GlobalsFunc) {
	var (
		incidentID string
		pathID     string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new escalation",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !shared.RequireFlag("incident", incidentID, "Provide --incident with the incident ID") {
				return nil
			}
			if !shared.RequireFlag("path", pathID, "Provide --path with the escalation path ID") {
				return nil
			}

			g := globals()
			return shared.WithClient(g.APIKey, g.Org, g.Timeout, func(ctx context.Context, client *api.Client) error {
				params := api.CreateEscalationParams{
					IncidentID:       incidentID,
					EscalationPathID: pathID,
				}
				escalation, err := client.CreateEscalation(ctx, params)
				if err != nil {
					return err
				}
				shared.WriteItem(escalation, g.Format)
				return nil
			})
		},
	}

	cmd.Flags().StringVar(&incidentID, "incident", "", "Incident ID (required)")
	cmd.Flags().StringVar(&pathID, "path", "", "Escalation path ID (required)")
	parent.AddCommand(cmd)
}

func registerPaths(parent *cobra.Command, globals shared.GlobalsFunc) {
	paths := &cobra.Command{
		Use:   "path",
		Short: "Manage escalation paths",
	}

	registerPathsList(paths, globals)
	registerPathsGet(paths, globals)

	parent.AddCommand(paths)
}

func registerPathsList(parent *cobra.Command, globals shared.GlobalsFunc) {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List escalation paths",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.APIKey, g.Org, g.Timeout, func(ctx context.Context, client *api.Client) error {
				paths, err := client.ListEscalationPaths(ctx)
				if err != nil {
					return err
				}
				shared.WritePaginatedList(shared.ToAnySlice(paths), nil, g.Format)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
}

func registerPathsGet(parent *cobra.Command, globals shared.GlobalsFunc) {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a single escalation path by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.APIKey, g.Org, g.Timeout, func(ctx context.Context, client *api.Client) error {
				path, err := client.GetEscalationPath(ctx, args[0])
				if err != nil {
					return err
				}
				shared.WriteItem(path, g.Format)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
}
