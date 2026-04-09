package schedules

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/cli/shared"
)

func TestSchedulesList(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]any{
			"schedules": []api.Schedule{
				{ID: "01HRDKWWNGX330JQ4J1PERJP8Y", Name: "Primary On-Call"},
				{ID: "01HRDKXJX3HFSN0R66E93FRX23", Name: "Secondary On-Call"},
			},
		})
	})

	root := shared.NewTestRoot(Register)
	root.SetArgs([]string{"schedule", "list"})

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
			"schedule": api.Schedule{ID: "01HRDKZPT5Q2PAMWN5R8WY23C3", Name: "Ops On-Call"},
		})
	})

	root := shared.NewTestRoot(Register)
	root.SetArgs([]string{"schedule", "get", "01HRDKZPT5Q2PAMWN5R8WY23C3"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/schedules/01HRDKZPT5Q2PAMWN5R8WY23C3" {
		t.Errorf("expected path /v2/schedules/01HRDKZPT5Q2PAMWN5R8WY23C3, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
}

func TestSchedulesEntries(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string][]string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = r.URL.Query()
		json.NewEncoder(w).Encode(map[string]any{
			"schedule_entries": map[string]any{
				"final": []api.ScheduleEntry{
					{
						RotationID:  "rot-1",
						Fingerprint: "abc123",
						User:        api.ScheduleEntryUser{ID: "01GF688HF8NVA17SNPGPYV5XNZ", Name: "Alice", Email: "alice@example.com", Role: "responder"},
						StartAt:     "2026-04-09T08:00:00Z",
						EndAt:       "2026-04-09T16:00:00Z",
					},
				},
				"overrides": []api.ScheduleEntry{},
				"scheduled": []api.ScheduleEntry{
					{
						RotationID:  "rot-1",
						Fingerprint: "abc123",
						User:        api.ScheduleEntryUser{ID: "01GF688HF8NVA17SNPGPYV5XNZ", Name: "Alice", Email: "alice@example.com", Role: "responder"},
						StartAt:     "2026-04-09T08:00:00Z",
						EndAt:       "2026-04-09T16:00:00Z",
					},
				},
			},
		})
	})

	root := shared.NewTestRoot(Register)
	root.SetArgs([]string{
		"schedule", "entries", "01HRDKWWNGX330JQ4J1PERJP8Y",
		"--from", "2026-04-09T08:00:00Z",
		"--to", "2026-04-09T16:00:00Z",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/schedule_entries" {
		t.Errorf("expected path /v2/schedule_entries, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
	if vals := gotQuery["schedule_id"]; len(vals) == 0 || vals[0] != "01HRDKWWNGX330JQ4J1PERJP8Y" {
		t.Errorf("expected schedule_id=01HRDKWWNGX330JQ4J1PERJP8Y, got %v", gotQuery["schedule_id"])
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
				ScheduleID: "01HRDKWWNGX330JQ4J1PERJP8Y",
				UserID:     "01GF688HF8NVA17SNPGPYV5XNZ",
				StartAt:    "2026-04-09T10:00:00Z",
				EndAt:      "2026-04-09T18:00:00Z",
			},
		})
	})

	root := shared.NewTestRoot(Register)
	root.SetArgs([]string{
		"schedule", "override", "01HRDKWWNGX330JQ4J1PERJP8Y",
		"--user", "01GF688HF8NVA17SNPGPYV5XNZ",
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
	if gotBody["schedule_id"] != "01HRDKWWNGX330JQ4J1PERJP8Y" {
		t.Errorf("expected schedule_id 01HRDKWWNGX330JQ4J1PERJP8Y, got %v", gotBody["schedule_id"])
	}
	if gotBody["user_id"] != "01GF688HF8NVA17SNPGPYV5XNZ" {
		t.Errorf("expected user_id 01GF688HF8NVA17SNPGPYV5XNZ, got %v", gotBody["user_id"])
	}
}
