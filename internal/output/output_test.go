package output

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	agenterrors "github.com/shhac/agent-incident/internal/errors"
)

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input   string
		want    Format
		wantErr bool
	}{
		{"json", FormatJSON, false},
		{"yaml", FormatYAML, false},
		{"jsonl", FormatNDJSON, false},
		{"ndjson", FormatNDJSON, false},
		{"xml", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseFormat(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %q", tt.input)
				}
				var apiErr *agenterrors.APIError
				if !agenterrors.As(err, &apiErr) {
					t.Fatal("expected APIError")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestResolveFormat(t *testing.T) {
	t.Run("empty flag returns default", func(t *testing.T) {
		got := ResolveFormat("", FormatYAML)
		if got != FormatYAML {
			t.Fatalf("expected %q, got %q", FormatYAML, got)
		}
	})

	t.Run("valid flag overrides default", func(t *testing.T) {
		got := ResolveFormat("yaml", FormatJSON)
		if got != FormatYAML {
			t.Fatalf("expected %q, got %q", FormatYAML, got)
		}
	})

	t.Run("invalid flag returns default", func(t *testing.T) {
		got := ResolveFormat("xml", FormatJSON)
		if got != FormatJSON {
			t.Fatalf("expected %q, got %q", FormatJSON, got)
		}
	})
}

func TestPruneNulls(t *testing.T) {
	t.Run("nested map", func(t *testing.T) {
		input := map[string]any{
			"a": "keep",
			"b": nil,
			"c": map[string]any{
				"d": nil,
				"e": "keep",
			},
		}
		result := pruneNulls(input).(map[string]any)
		if _, ok := result["b"]; ok {
			t.Fatal("expected b to be pruned")
		}
		inner := result["c"].(map[string]any)
		if _, ok := inner["d"]; ok {
			t.Fatal("expected c.d to be pruned")
		}
		if inner["e"] != "keep" {
			t.Fatal("expected c.e to be kept")
		}
	})

	t.Run("array with nulls inside maps", func(t *testing.T) {
		input := []any{
			map[string]any{"x": nil, "y": float64(1)},
			"plain",
		}
		result := pruneNulls(input).([]any)
		if len(result) != 2 {
			t.Fatalf("expected 2 elements, got %d", len(result))
		}
		m := result[0].(map[string]any)
		if _, ok := m["x"]; ok {
			t.Fatal("expected x to be pruned")
		}
		if m["y"] != float64(1) {
			t.Fatal("expected y to be kept")
		}
	})

	t.Run("scalar passthrough", func(t *testing.T) {
		if pruneNulls("hello") != "hello" {
			t.Fatal("scalar should pass through unchanged")
		}
		if pruneNulls(float64(42)) != float64(42) {
			t.Fatal("number should pass through unchanged")
		}
	})
}

func TestWriteError(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		var buf bytes.Buffer
		err := agenterrors.New("auth failed", agenterrors.FixableByHuman).WithHint("check key")
		WriteError(&buf, err)
		var result map[string]any
		if e := json.Unmarshal(buf.Bytes(), &result); e != nil {
			t.Fatalf("invalid JSON: %v", e)
		}
		if result["error"] != "auth failed" {
			t.Fatalf("expected error %q, got %v", "auth failed", result["error"])
		}
		if result["hint"] != "check key" {
			t.Fatalf("expected hint %q, got %v", "check key", result["hint"])
		}
		if result["fixable_by"] != "human" {
			t.Fatalf("expected fixable_by %q, got %v", "human", result["fixable_by"])
		}
	})

	t.Run("without hint", func(t *testing.T) {
		var buf bytes.Buffer
		err := agenterrors.New("bad request", agenterrors.FixableByAgent)
		WriteError(&buf, err)
		var result map[string]any
		if e := json.Unmarshal(buf.Bytes(), &result); e != nil {
			t.Fatalf("invalid JSON: %v", e)
		}
		if _, ok := result["hint"]; ok {
			t.Fatal("expected no hint key")
		}
	})

	t.Run("plain error gets wrapped", func(t *testing.T) {
		var buf bytes.Buffer
		WriteError(&buf, &plainError{msg: "wrapped plain"})
		var result map[string]any
		if e := json.Unmarshal(buf.Bytes(), &result); e != nil {
			t.Fatalf("invalid JSON: %v", e)
		}
		if result["fixable_by"] != "agent" {
			t.Fatalf("expected fixable_by agent, got %v", result["fixable_by"])
		}
	})
}

func TestWriteErrorPlainError(t *testing.T) {
	var buf bytes.Buffer
	plainErr := &plainError{msg: "something happened"}
	WriteError(&buf, plainErr)
	var result map[string]any
	if e := json.Unmarshal(buf.Bytes(), &result); e != nil {
		t.Fatalf("invalid JSON: %v", e)
	}
	if result["error"] != "something happened" {
		t.Fatalf("expected %q, got %v", "something happened", result["error"])
	}
	if result["fixable_by"] != "agent" {
		t.Fatalf("expected fixable_by %q, got %v", "agent", result["fixable_by"])
	}
}

type plainError struct{ msg string }

func (e *plainError) Error() string { return e.msg }

func TestNDJSONWriter(t *testing.T) {
	t.Run("WriteItem", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewNDJSONWriter(&buf)
		if err := w.WriteItem(map[string]string{"name": "alice"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := w.WriteItem(map[string]string{"name": "bob"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 2 {
			t.Fatalf("expected 2 lines, got %d", len(lines))
		}
		var first map[string]string
		if err := json.Unmarshal([]byte(lines[0]), &first); err != nil {
			t.Fatalf("invalid JSON line: %v", err)
		}
		if first["name"] != "alice" {
			t.Fatalf("expected alice, got %s", first["name"])
		}
	})

	t.Run("WritePagination", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewNDJSONWriter(&buf)
		err := w.WritePagination(&Pagination{
			HasMore:    true,
			TotalItems: 100,
			NextCursor: "abc123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var result map[string]any
		if e := json.Unmarshal(buf.Bytes(), &result); e != nil {
			t.Fatalf("invalid JSON: %v", e)
		}
		pag := result["@pagination"].(map[string]any)
		if pag["has_more"] != true {
			t.Fatal("expected has_more true")
		}
		if pag["next_cursor"] != "abc123" {
			t.Fatalf("expected next_cursor abc123, got %v", pag["next_cursor"])
		}
	})
}

func TestPrintCaptureStdout(t *testing.T) {
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	Print(map[string]string{"key": "value"}, FormatJSON, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, `"key": "value"`) {
		t.Fatalf("expected JSON output with key/value, got: %s", output)
	}
}
