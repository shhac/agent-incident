package schedules

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/cli/shared"
)

// Register adds the schedules command group to the root command.
func Register(root *cobra.Command, globals shared.GlobalsFunc) {
	schedules := &cobra.Command{
		Use:   "schedules",
		Short: "Manage on-call schedules",
	}

	registerList(schedules, globals)
	registerGet(schedules, globals)
	registerEntries(schedules, globals)
	registerOverride(schedules, globals)
	registerLLMHelp(schedules)

	root.AddCommand(schedules)
}

func registerList(parent *cobra.Command, globals shared.GlobalsFunc) {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List schedules",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.APIKey, g.Org, g.Timeout, func(ctx context.Context, client *api.Client) error {
				schedules, err := client.ListSchedules(ctx)
				if err != nil {
					return err
				}
				shared.WritePaginatedList(shared.ToAnySlice(schedules), nil, g.Format)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
}

func registerGet(parent *cobra.Command, globals shared.GlobalsFunc) {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a single schedule by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.APIKey, g.Org, g.Timeout, func(ctx context.Context, client *api.Client) error {
				schedule, err := client.GetSchedule(ctx, args[0])
				if err != nil {
					return err
				}
				shared.WriteItem(schedule, g.Format)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
}

func registerEntries(parent *cobra.Command, globals shared.GlobalsFunc) {
	var (
		from string
		to   string
	)

	cmd := &cobra.Command{
		Use:   "entries <schedule-id>",
		Short: "List schedule entries for a time window",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fromTime, toTime, err := shared.ParseTimeRange(from, to)
			if err != nil {
				return err
			}

			g := globals()
			return shared.WithClient(g.APIKey, g.Org, g.Timeout, func(ctx context.Context, client *api.Client) error {
				entries, err := client.ListScheduleEntries(ctx, args[0], fromTime, toTime)
				if err != nil {
					return err
				}
				shared.WritePaginatedList(shared.ToAnySlice(entries), nil, g.Format)
				return nil
			})
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "Window start (relative, RFC3339, or epoch; default: now-1h)")
	cmd.Flags().StringVar(&to, "to", "", "Window end (relative, RFC3339, or epoch; default: now)")
	parent.AddCommand(cmd)
}

func registerOverride(parent *cobra.Command, globals shared.GlobalsFunc) {
	var (
		userID string
		from   string
		to     string
	)

	cmd := &cobra.Command{
		Use:   "override <schedule-id>",
		Short: "Create a schedule override",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !shared.RequireFlag("user", userID, "Provide --user with the user ID for the override") {
				return nil
			}
			if !shared.RequireFlag("from", from, "Provide --from for the override start time") {
				return nil
			}
			if !shared.RequireFlag("to", to, "Provide --to for the override end time") {
				return nil
			}

			fromTime, err := shared.ParseTime(from)
			if err != nil {
				return err
			}
			toTime, err := shared.ParseTime(to)
			if err != nil {
				return err
			}

			g := globals()
			return shared.WithClient(g.APIKey, g.Org, g.Timeout, func(ctx context.Context, client *api.Client) error {
				params := api.ScheduleOverrideParams{
					ScheduleID: args[0],
					UserID:     userID,
					StartAt:    fromTime.Format("2006-01-02T15:04:05Z07:00"),
					EndAt:      toTime.Format("2006-01-02T15:04:05Z07:00"),
				}
				override, err := client.CreateScheduleOverride(ctx, params)
				if err != nil {
					return err
				}
				shared.WriteItem(override, g.Format)
				return nil
			})
		},
	}

	cmd.Flags().StringVar(&userID, "user", "", "User ID for the override (required)")
	cmd.Flags().StringVar(&from, "from", "", "Override start time (required)")
	cmd.Flags().StringVar(&to, "to", "", "Override end time (required)")
	parent.AddCommand(cmd)
}
