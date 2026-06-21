package shared

import (
	"context"
	"os"

	"github.com/shhac/agent-incident/internal/api"
	agenterrors "github.com/shhac/agent-incident/internal/errors"
	"github.com/shhac/agent-incident/internal/output"
	libcli "github.com/shhac/lib-agent-cli/cli"
)

// GlobalFlags holds persistent flags available to all commands. The shared
// --format/--timeout/--debug live in the embedded libcli.Globals; --org and
// --api-key are incident.io domain flags.
type GlobalFlags struct {
	libcli.Globals // Format, TimeoutMS, Debug

	Org    string
	APIKey string
}

// GlobalsFunc is the signature for the globals accessor passed to domain Register functions.
type GlobalsFunc = func() *GlobalFlags

func ToAnySlice[T any](s []T) []any {
	result := make([]any, len(s))
	for i, v := range s {
		result[i] = v
	}
	return result
}

// WritePaginatedList writes a list in the resolved format (default: NDJSON).
func WritePaginatedList(items []any, pagination *output.Pagination, format string) {
	f := output.ResolveFormat(format, output.FormatNDJSON)
	if f == output.FormatNDJSON {
		w := output.NewNDJSONWriter(os.Stdout)
		for _, item := range items {
			_ = w.WriteItem(item)
		}
		if pagination != nil {
			_ = w.WritePagination(pagination)
		}
		return
	}
	result := map[string]any{"data": items}
	if pagination != nil {
		result["pagination"] = pagination
	}
	output.Print(result, f, true)
}

// WriteItem writes a single item in the resolved format (default: JSON).
func WriteItem(data any, format string) {
	f := output.ResolveFormat(format, output.FormatJSON)
	output.Print(data, f, true)
}

// CursorPagination builds pagination metadata from a cursor string.
func CursorPagination(cursor string) *output.Pagination {
	if cursor == "" {
		return nil
	}
	return &output.Pagination{HasMore: true, NextCursor: cursor}
}

// GetEntities runs the family's multi-capable get for the incident domain: it
// sets up one client, then resolves each id through getOne and streams the
// result per the shared get contract (NDJSON by default — one record or
// {"@unresolved":…} per id in input order; item-level misses stay on stdout,
// command-level failures bubble to the single sink). getOne returns the record
// for an id, or a classified *agenterrors.APIError (a lib-agent-output *Error)
// so a 404/bad input becomes an @unresolved record rather than aborting.
func GetEntities(g *GlobalFlags, args []string, getOne func(ctx context.Context, client *api.Client, id string) (any, error)) error {
	return WithClient(g, func(ctx context.Context, client *api.Client) error {
		return libcli.EntityGet(os.Stdout, g.Format, args, func(id string) (any, error) {
			return getOne(ctx, client, id)
		})
	})
}

// RequireFlag returns a structured fixable_by:agent error when a flag value is
// empty, and nil otherwise. Callers return the error so the root Run sink
// renders it once and exits non-zero — they must not render it themselves.
func RequireFlag(flag, value, hint string) error {
	if value != "" {
		return nil
	}
	err := agenterrors.Newf(agenterrors.FixableByAgent, "--%s is required", flag)
	if hint != "" {
		err = err.WithHint(hint)
	}
	return err
}
