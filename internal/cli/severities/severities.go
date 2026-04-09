package severities

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/cli/shared"
)

func Register(root *cobra.Command, globals shared.GlobalsFunc) {
	sev := &cobra.Command{
		Use:   "severities",
		Short: "List and inspect incident severities",
	}

	registerList(sev, globals)
	registerGet(sev, globals)
	shared.RegisterLLMHelp(sev, "LLM reference for severities", llmHelpText)

	root.AddCommand(sev)
}

func registerList(parent *cobra.Command, globals shared.GlobalsFunc) {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all severities",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.APIKey, g.Org, g.Timeout, func(ctx context.Context, client *api.Client) error {
				items, err := client.ListSeverities(ctx)
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
		Use:   "get <id>",
		Short: "Get a severity by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.APIKey, g.Org, g.Timeout, func(ctx context.Context, client *api.Client) error {
				item, err := client.GetSeverity(ctx, args[0])
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
