package schedules

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/cli/shared"
)

func newTestRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "agent-incident",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	globals := func() *shared.GlobalFlags {
		return &shared.GlobalFlags{}
	}
	Register(root, globals)
	return root
}

func TestSchedulesList(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]any{
			"schedules": []api.Schedule{
				{ID: "sched-1", Name: "Primary On-Call"},
				{ID: "sched-2", Name: "Secondary On-Call"},
			},
		})
	})

	root := newTestRoot()
	root.SetArgs([]string{"schedules", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/schedules" {
		t.Errorf("expected path /v2/schedules, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
}

func TestSchedulesGet(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]any{
			"schedule": api.Schedule{ID: "sched-42", Name: "Ops On-Call"},
		})
	})

	root := newTestRoot()
	root.SetArgs([]string{"schedules", "get", "sched-42"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/schedules/sched-42" {
		t.Errorf("expected path /v2/schedules/sched-42, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
}

func TestSchedulesOverride(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody map[string]any

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		if r.Method == http.MethodPost {
			json.NewDecoder(r.Body).Decode(&gotBody)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"schedule_override": api.ScheduleOverride{
				ID:         "ovr-1",
				ScheduleID: "sched-1",
				UserID:     "user-1",
				StartAt:    "2026-04-09T10:00:00Z",
				EndAt:      "2026-04-09T18:00:00Z",
			},
		})
	})

	root := newTestRoot()
	root.SetArgs([]string{
		"schedules", "override", "sched-1",
		"--user", "user-1",
		"--from", "2026-04-09T10:00:00Z",
		"--to", "2026-04-09T18:00:00Z",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/schedule_overrides" {
		t.Errorf("expected path /v2/schedule_overrides, got %q", gotPath)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", gotMethod)
	}
	if gotBody["schedule_id"] != "sched-1" {
		t.Errorf("expected schedule_id sched-1, got %v", gotBody["schedule_id"])
	}
	if gotBody["user_id"] != "user-1" {
		t.Errorf("expected user_id user-1, got %v", gotBody["user_id"])
	}
}
