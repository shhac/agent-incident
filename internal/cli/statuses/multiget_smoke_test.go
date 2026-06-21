package statuses

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/cli/shared"
)

// TestStatusesGetMulti verifies: single get → one NDJSON line;
// get(good, bogus, good) → record, @unresolved, record (exit 0).
func TestStatusesGetMulti(t *testing.T) {
	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/v1/incident_statuses/")
		if id == "bogus-id" {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"incident_status": api.IncidentStatusResource{ID: id, Name: "Monitoring", Category: "active", Rank: 1},
		})
	})

	t.Run("single get NDJSON", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		root := shared.NewTestRoot(Register)
		root.SetArgs([]string{"status", "get", "st-1"})
		err := root.Execute()

		_ = w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		out := strings.TrimSpace(buf.String())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		lines := strings.Split(out, "\n")
		if len(lines) != 1 {
			t.Fatalf("expected 1 NDJSON line, got %d: %s", len(lines), out)
		}
		if !strings.Contains(lines[0], `"st-1"`) {
			t.Errorf("expected st-1 in output, got: %s", lines[0])
		}
	})

	t.Run("multi get: record, @unresolved, record (exit 0)", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		root := shared.NewTestRoot(Register)
		root.SetArgs([]string{"status", "get", "st-1", "bogus-id", "st-2"})
		err := root.Execute()

		_ = w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		out := strings.TrimSpace(buf.String())

		if err != nil {
			t.Fatalf("expected exit 0 for per-item miss, got error: %v", err)
		}
		lines := strings.Split(out, "\n")
		if len(lines) != 3 {
			t.Fatalf("expected 3 lines, got %d:\n%s", len(lines), out)
		}
		if strings.Contains(lines[0], "@unresolved") || !strings.Contains(lines[0], `"st-1"`) {
			t.Errorf("line 1: expected record for st-1, got: %s", lines[0])
		}
		if !strings.Contains(lines[1], `"@unresolved"`) || !strings.Contains(lines[1], "bogus-id") {
			t.Errorf("line 2: expected @unresolved for bogus-id, got: %s", lines[1])
		}
		if strings.Contains(lines[2], "@unresolved") || !strings.Contains(lines[2], `"st-2"`) {
			t.Errorf("line 3: expected record for st-2, got: %s", lines[2])
		}
	})
}
