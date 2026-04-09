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
		Use:   "schedule",
		Short: "Manage on-call schedules",
	}

	registerList(schedules, globals)
	registerGet(schedules, globals)
	registerEntries(schedules, globals)
	registerOverride(schedules, globals)
	shared.RegisterLLMHelp(schedules, "LLM reference for schedules commands", schedulesLLMHelp)

	root.AddCommand(schedules)
}

func registerList(parent *cobra.Command, globals shared.GlobalsFunc) {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List schedules",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g, func(ctx context.Context, client *api.Client) error {
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
		Use:   "get <name-or-id>",
		Short: "Get a single schedule by name or ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g, func(ctx context.Context, client *api.Client) error {
				id, err := shared.ResolveScheduleID(ctx, client, args[0])
				if err != nil {
					return err
				}
				schedule, err := client.GetSchedule(ctx, id)
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
		Use:   "entries <name-or-id>",
		Short: "List schedule entries for a time window",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fromTime, toTime, err := shared.ParseTimeRange(from, to)
			if err != nil {
				return err
			}

			g := globals()
			return shared.WithClient(g, func(ctx context.Context, client *api.Client) error {
				id, err := shared.ResolveScheduleID(ctx, client, args[0])
				if err != nil {
					return err
				}
				entries, err := client.ListScheduleEntries(ctx, id, fromTime, toTime)
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
		Use:   "override <name-or-id>",
		Short: "Create a schedule override",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !shared.RequireFlag("user", userID, "Provide --user with the user name, email, or ID for the override") {
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
			return shared.WithClient(g, func(ctx context.Context, client *api.Client) error {
				scheduleID, err := shared.ResolveScheduleID(ctx, client, args[0])
				if err != nil {
					return err
				}
				resolvedUserID, err := shared.ResolveUserID(ctx, client, userID)
				if err != nil {
					return err
				}
				params := api.ScheduleOverrideParams{
					ScheduleID: scheduleID,
					UserID:     resolvedUserID,
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

	cmd.Flags().StringVar(&userID, "user", "", "User name, email, or ID for the override (required)")
	cmd.Flags().StringVar(&from, "from", "", "Override start time (required)")
	cmd.Flags().StringVar(&to, "to", "", "Override end time (required)")
	parent.AddCommand(cmd)
}
