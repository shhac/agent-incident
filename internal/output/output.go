package output

import (
	"encoding/json"
	"io"
	"os"

	agenterrors "github.com/shhac/agent-incident/internal/errors"
	"gopkg.in/yaml.v3"
)

type Format string

const (
	FormatJSON   Format = "json"
	FormatYAML   Format = "yaml"
	FormatNDJSON Format = "jsonl"
)

func ParseFormat(s string) (Format, error) {
	switch s {
	case "json":
		return FormatJSON, nil
	case "yaml":
		return FormatYAML, nil
	case "jsonl", "ndjson":
		return FormatNDJSON, nil
	default:
		return "", agenterrors.Newf(agenterrors.FixableByAgent, "unknown format %q, expected: json, yaml, jsonl", s)
	}
}

func ResolveFormat(flagFormat string, defaultFormat Format) Format {
	if flagFormat == "" {
		return defaultFormat
	}
	f, err := ParseFormat(flagFormat)
	if err != nil {
		return defaultFormat
	}
	return f
}

func Print(data any, format Format, prune bool) {
	switch format {
	case FormatYAML:
		printYAML(data, prune)
	default:
		printJSON(data, prune)
	}
}

func PrintJSON(data any, prune bool) {
	printJSON(data, prune)
}

// toCleanAny round-trips data through JSON to get a generic any value,
// optionally pruning null fields.
func toCleanAny(data any, prune bool) (any, bool) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, false
	}
	var decoded any
	if err := json.Unmarshal(b, &decoded); err != nil {
		return nil, false
	}
	if prune {
		decoded = pruneNulls(decoded)
	}
	return decoded, true
}

func printJSON(data any, prune bool) {
	cleaned, ok := toCleanAny(data, prune)
	if !ok {
		return
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	_ = enc.Encode(cleaned)
}

func printYAML(data any, prune bool) {
	cleaned, ok := toCleanAny(data, prune)
	if !ok {
		return
	}
	enc := yaml.NewEncoder(os.Stdout)
	enc.SetIndent(2)
	_ = enc.Encode(cleaned)
}

func WriteError(w io.Writer, err error) {
	var aerr *agenterrors.APIError
	if !agenterrors.As(err, &aerr) {
		aerr = agenterrors.Wrap(err, agenterrors.FixableByAgent)
	}
	payload := map[string]any{
		"error":      aerr.Message,
		"fixable_by": string(aerr.FixableBy),
	}
	if aerr.Hint != "" {
		payload["hint"] = aerr.Hint
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(payload)
}

// NDJSONWriter writes one JSON object per line.
type NDJSONWriter struct {
	enc *json.Encoder
}

func NewNDJSONWriter(w io.Writer) *NDJSONWriter {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return &NDJSONWriter{enc: enc}
}

func (n *NDJSONWriter) WriteItem(item any) error {
	return n.enc.Encode(item)
}

// Pagination metadata for cursor-based APIs.
type Pagination struct {
	HasMore    bool   `json:"has_more"`
	TotalItems int    `json:"total_items,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
}

func (n *NDJSONWriter) WritePagination(p *Pagination) error {
	return n.enc.Encode(map[string]any{"@pagination": p})
}

func pruneNulls(v any) any {
	switch val := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, v := range val {
			if v == nil {
				continue
			}
			out[k] = pruneNulls(v)
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, v := range val {
			out[i] = pruneNulls(v)
		}
		return out
	default:
		return v
	}
}
