package incidents

import (
	"context"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/cli/shared"
)

// normalizeIncidentRef strips a "INC-" prefix when followed by digits,
// so users can pass "INC-2000" or "2000" as an incident reference.
// UUIDs and other ID formats are passed through unchanged.
func normalizeIncidentRef(ref string) string {
	upper := strings.ToUpper(ref)
	if strings.HasPrefix(upper, "INC-") {
		suffix := ref[4:]
		if _, err := strconv.Atoi(suffix); err == nil {
			return suffix
		}
	}
	return ref
}

// Register adds the incidents command group to the root command.
func Register(root *cobra.Command, globals shared.GlobalsFunc) {
	incidents := &cobra.Command{
		Use:   "incident",
		Short: "Manage incidents",
	}

	registerList(incidents, globals)
	registerGet(incidents, globals)
	registerCreate(incidents, globals)
	registerEdit(incidents, globals)
	registerUpdates(incidents, globals)
	shared.RegisterLLMHelp(incidents, "LLM reference for incidents commands", incidentsLLMHelp)

	root.AddCommand(incidents)
}

func registerList(parent *cobra.Command, globals shared.GlobalsFunc) {
	var (
		status   []string
		severity []string
		since    string
		until    string
		limit    int
		after    string
		full     bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List incidents",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()

			var createdAfter, createdBefore string
			if since != "" {
				t, err := shared.ParseTime(since)
				if err != nil {
					return err
				}
				createdAfter = t.Format("2006-01-02")
			}
			if until != "" {
				t, err := shared.ParseTime(until)
				if err != nil {
					return err
				}
				createdBefore = t.Format("2006-01-02")
			}

			return shared.WithClient(g, func(ctx context.Context, client *api.Client) error {
				opts := api.ListIncidentsOpts{
					StatusCategory: status,
					Severity:       severity,
					CreatedAfter:   createdAfter,
					CreatedBefore:  createdBefore,
					PageSize:       limit,
					After:          after,
				}

				incidents, cursor, err := client.ListIncidents(ctx, opts)
				if err != nil {
					return err
				}

				pagination := shared.CursorPagination(cursor)

				if full {
					shared.WritePaginatedList(shared.ToAnySlice(incidents), pagination, g.Format)
				} else {
					compact := api.ToCompact(incidents)
					shared.WritePaginatedList(shared.ToAnySlice(compact), pagination, g.Format)
				}
				return nil
			})
		},
	}

	cmd.Flags().StringSliceVar(&status, "status", nil, "Filter by status category (active, closed, etc.)")
	cmd.Flags().StringSliceVar(&severity, "severity", nil, "Filter by severity name")
	cmd.Flags().StringVar(&since, "since", "", "Only show incidents created after this time")
	cmd.Flags().StringVar(&until, "until", "", "Only show incidents created before this time")
	cmd.Flags().IntVar(&limit, "limit", 25, "Number of results per page")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor")
	cmd.Flags().BoolVar(&full, "full", false, "Return full incident objects instead of compact")
	parent.AddCommand(cmd)
}

func registerGet(parent *cobra.Command, globals shared.GlobalsFunc) {
	cmd := &cobra.Command{
		Use:   "get <id-or-reference>",
		Short: "Get a single incident by ID or reference (e.g. INC-2000 or 2000)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g, func(ctx context.Context, client *api.Client) error {
				id := normalizeIncidentRef(args[0])
				incident, err := client.GetIncident(ctx, id)
				if err != nil {
					return err
				}
				shared.WriteItem(incident, g.Format)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
}

func registerCreate(parent *cobra.Command, globals shared.GlobalsFunc) {
	var (
		name     string
		severity string
		summary  string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new incident",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !shared.RequireFlag("name", name, "Provide --name for the incident") {
				return nil
			}
			g := globals()
			return shared.WithClient(g, func(ctx context.Context, client *api.Client) error {
				params := api.CreateIncidentParams{
					Name:           name,
					Summary:        summary,
					SeverityID:     severity,
					IdempotencyKey: shared.NewIdempotencyKey(),
				}
				incident, err := client.CreateIncident(ctx, params)
				if err != nil {
					return err
				}
				shared.WriteItem(incident, g.Format)
				return nil
			})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Incident name (required)")
	cmd.Flags().StringVar(&severity, "severity", "", "Severity ID")
	cmd.Flags().StringVar(&summary, "summary", "", "Incident summary")
	parent.AddCommand(cmd)
}

func registerEdit(parent *cobra.Command, globals shared.GlobalsFunc) {
	var (
		name     string
		severity string
		summary  string
	)

	cmd := &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit an existing incident",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g, func(ctx context.Context, client *api.Client) error {
				fields := api.EditIncidentFields{}
				if cmd.Flags().Changed("name") {
					fields.Name = &name
				}
				if cmd.Flags().Changed("severity") {
					fields.SeverityID = &severity
				}
				if cmd.Flags().Changed("summary") {
					fields.Summary = &summary
				}

				params := api.EditIncidentParams{Incident: fields}
				incident, err := client.EditIncident(ctx, args[0], params)
				if err != nil {
					return err
				}
				shared.WriteItem(incident, g.Format)
				return nil
			})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "New incident name")
	cmd.Flags().StringVar(&severity, "severity", "", "New severity ID")
	cmd.Flags().StringVar(&summary, "summary", "", "Updated summary")
	parent.AddCommand(cmd)
}

func registerUpdates(parent *cobra.Command, globals shared.GlobalsFunc) {
	var (
		limit int
		after string
	)

	cmd := &cobra.Command{
		Use:   "updates <id>",
		Short: "List updates for an incident",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g, func(ctx context.Context, client *api.Client) error {
				updates, cursor, err := client.ListIncidentUpdates(ctx, args[0], limit, after)
				if err != nil {
					return err
				}
				pagination := shared.CursorPagination(cursor)
				shared.WritePaginatedList(shared.ToAnySlice(updates), pagination, g.Format)
				return nil
			})
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 25, "Number of results per page")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor")
	parent.AddCommand(cmd)
}

