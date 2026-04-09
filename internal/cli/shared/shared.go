package shared

import (
	"os"

	agenterrors "github.com/shhac/agent-incident/internal/errors"
	"github.com/shhac/agent-incident/internal/output"
)

// GlobalFlags holds persistent flags available to all commands.
type GlobalFlags struct {
	Org     string
	APIKey  string
	Format  string
	Timeout int
	Debug   bool
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

// RequireFlag checks that a flag value is non-empty, writing an error to stderr if not.
func RequireFlag(flag, value, hint string) bool {
	if value != "" {
		return true
	}
	err := agenterrors.Newf(agenterrors.FixableByAgent, "--%s is required", flag)
	if hint != "" {
		err = err.WithHint(hint)
	}
	output.WriteError(os.Stderr, err)
	return false
}
