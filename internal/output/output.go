// Package output re-exports the shared output contract from lib-agent-output,
// keeping the internal/output import path while the wire mechanism (format
// parsing, JSON/YAML encoding, error rendering) lives in one place. What stays
// local is agent-incident policy: the single-return ResolveFormat the commands
// call, the Print(data, format, prune bool) signature, and the incident.io-shaped
// pagination trailer. (Migration shim.)
package output

import (
	"bytes"
	"io"
	"os"

	out "github.com/shhac/lib-agent-output"
	"gopkg.in/yaml.v3"
)

// Format and its values come from the shared contract; ParseFormat is therefore
// the family's lenient parser (accepts "ndjson"/"yml", case-insensitive).
type Format = out.Format

const (
	FormatJSON   = out.FormatJSON
	FormatYAML   = out.FormatYAML
	FormatNDJSON = out.FormatNDJSON
)

var (
	ParseFormat = out.ParseFormat
	WriteError  = out.WriteError
)

// init registers agent-incident's YAML encoder with lib-agent-output, so YAML
// support (and its yaml.v3 dependency) stays in this CLI while the core library
// remains dependency-free.
func init() {
	out.RegisterEncoder(out.FormatYAML, func(v any) ([]byte, error) {
		var buf bytes.Buffer
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		if err := enc.Encode(v); err != nil {
			return nil, err
		}
		_ = enc.Close()
		return buf.Bytes(), nil
	})
}

// ResolveFormat keeps agent-incident's single-return signature (the commands
// call it without error handling): a bad flag falls back to the default rather
// than surfacing a parse error.
func ResolveFormat(flagFormat string, defaultFormat Format) Format {
	f, err := out.ResolveFormat(flagFormat, defaultFormat)
	if err != nil {
		return defaultFormat
	}
	return f
}

// Print encodes data in the given format via the shared encoder, optionally
// pruning nil fields first.
func Print(data any, format Format, prune bool) {
	_ = out.Print(os.Stdout, data, format, pruner(prune))
}

// PrintJSON encodes data as pretty JSON, optionally pruning nil fields.
func PrintJSON(data any, prune bool) {
	Print(data, FormatJSON, prune)
}

func pruner(prune bool) out.Pruner {
	if prune {
		return out.PruneNils
	}
	return nil
}

// pruneNulls drops nil map values recursively. Retained as the local name the
// package's tests pin; it delegates to the shared PruneNils policy.
func pruneNulls(v any) any { return out.PruneNils(v) }

// NDJSONWriter writes one JSON object per line.
type NDJSONWriter struct {
	w *out.NDJSONWriter
}

func NewNDJSONWriter(w io.Writer) *NDJSONWriter {
	return &NDJSONWriter{w: out.NewNDJSONWriter(w)}
}

func (n *NDJSONWriter) WriteItem(item any) error {
	return n.w.WriteItem(item)
}

// Pagination metadata for cursor-based APIs. incident.io exposes a next cursor
// and total count, so it stays local rather than using out.Pagination's value
// receiver — the commands hold a *Pagination.
type Pagination struct {
	HasMore    bool   `json:"has_more"`
	TotalItems int    `json:"total_items,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
}

func (n *NDJSONWriter) WritePagination(p *Pagination) error {
	return n.w.WriteMetaLine(out.MetaKeyPagination, p)
}
