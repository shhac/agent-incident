package timestamps

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/shhac/agent-incident/internal/cli/shared"
)

func TestTimestampsList(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]any{
			"incident_timestamps": []map[string]any{
				{"id": "ts-1", "name": "Reported at", "rank": 0},
				{"id": "ts-2", "name": "Resolved at", "rank": 5},
			},
		})
	})

	root := shared.NewTestRoot(Register)
	root.SetArgs([]string{"timestamp", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/incident_timestamps" {
		t.Errorf("expected path /v2/incident_timestamps, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
}

func TestTimestampsGet(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]any{
			"incident_timestamp": map[string]any{
				"id": "ts-1", "name": "Reported at", "rank": 0,
			},
		})
	})

	root := shared.NewTestRoot(Register)
	root.SetArgs([]string{"timestamp", "get", "ts-1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/incident_timestamps/ts-1" {
		t.Errorf("expected path /v2/incident_timestamps/ts-1, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
}
