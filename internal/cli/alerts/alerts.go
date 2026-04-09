package alerts

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/cli/shared"
)

// Register adds the alerts command group to the root command.
func Register(root *cobra.Command, globals shared.GlobalsFunc) {
	alerts := &cobra.Command{
		Use:   "alerts",
		Short: "Manage alerts and alert events",
	}

	registerList(alerts, globals)
	registerGet(alerts, globals)
	registerCreate(alerts, globals)
	registerIncidents(alerts, globals)
	shared.RegisterLLMHelp(alerts, "LLM reference for alerts commands", alertsLLMHelp)

	root.AddCommand(alerts)
}

func registerList(parent *cobra.Command, globals shared.GlobalsFunc) {
	var (
		status string
		source string
		limit  int
		after  string
		full   bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List alerts with optional filters",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.APIKey, g.Org, g.Timeout, func(ctx context.Context, client *api.Client) error {
				opts := api.ListAlertsOpts{
					PageSize: limit,
					After:    after,
				}

				if status != "" {
					opts.Status = strings.Split(status, ",")
				}
				if source != "" {
					opts.DeduplicationKey = source
				}

				result, err := client.ListAlerts(ctx, opts)
				if err != nil {
					return err
				}

				pagination := shared.CursorPagination(result.After)

				if full {
					shared.WritePaginatedList(shared.ToAnySlice(result.Alerts), pagination, g.Format)
					return nil
				}

				compact := api.ToCompactAlerts(result.Alerts)
				shared.WritePaginatedList(shared.ToAnySlice(compact), pagination, g.Format)
				return nil
			})
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "Filter by status (comma-separated, e.g. firing,resolved)")
	cmd.Flags().StringVar(&source, "source", "", "Filter by deduplication key")
	cmd.Flags().IntVar(&limit, "limit", 25, "Maximum number of alerts to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor")
	cmd.Flags().BoolVar(&full, "full", false, "Show full alert details instead of compact view")

	parent.AddCommand(cmd)
}

func registerGet(parent *cobra.Command, globals shared.GlobalsFunc) {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a single alert by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.APIKey, g.Org, g.Timeout, func(ctx context.Context, client *api.Client) error {
				alert, err := client.GetAlert(ctx, args[0])
				if err != nil {
					return err
				}
				shared.WriteItem(alert, g.Format)
				return nil
			})
		},
	}

	parent.AddCommand(cmd)
}

func registerCreate(parent *cobra.Command, globals shared.GlobalsFunc) {
	var (
		sourceID    string
		title       string
		description string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an alert event via an HTTP alert source",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !shared.RequireFlag("source-id", sourceID, "Provide --source-id <alert_source_config_id>") {
				return nil
			}
			if !shared.RequireFlag("title", title, "Provide --title <alert title>") {
				return nil
			}

			g := globals()
			return shared.WithClient(g.APIKey, g.Org, g.Timeout, func(ctx context.Context, client *api.Client) error {
				params := api.CreateAlertEventParams{
					Title:       title,
					Description: description,
				}
				alert, err := client.CreateAlertEvent(ctx, sourceID, params)
				if err != nil {
					return err
				}
				shared.WriteItem(alert, g.Format)
				return nil
			})
		},
	}

	cmd.Flags().StringVar(&sourceID, "source-id", "", "Alert source config ID (required)")
	cmd.Flags().StringVar(&title, "title", "", "Alert title (required)")
	cmd.Flags().StringVar(&description, "description", "", "Alert description")

	parent.AddCommand(cmd)
}

func registerIncidents(parent *cobra.Command, globals shared.GlobalsFunc) {
	var (
		limit int
		after string
	)

	cmd := &cobra.Command{
		Use:   "incidents",
		Short: "List alerts attached to incidents",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.APIKey, g.Org, g.Timeout, func(ctx context.Context, client *api.Client) error {
				result, err := client.ListIncidentAlerts(ctx, limit, after)
				if err != nil {
					return err
				}

				pagination := shared.CursorPagination(result.After)
				shared.WritePaginatedList(shared.ToAnySlice(result.IncidentAlerts), pagination, g.Format)
				return nil
			})
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 25, "Maximum number of results to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor")

	parent.AddCommand(cmd)
}
