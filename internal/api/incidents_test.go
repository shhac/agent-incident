package api

import (
	"strings"
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name string
		s    string
		max  int
		want string
	}{
		{name: "shorter than max", s: "hello", max: 10, want: "hello"},
		{name: "equal to max", s: "hello", max: 5, want: "hello"},
		{name: "longer than max", s: "hello world", max: 5, want: "hello"},
		{name: "empty string", s: "", max: 5, want: ""},
		{name: "max zero", s: "hello", max: 0, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.s, tt.max)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.s, tt.max, got, tt.want)
			}
		})
	}
}

func TestToCompact(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		result := ToCompact(nil)
		if len(result) != 0 {
			t.Fatalf("expected empty slice, got %d items", len(result))
		}
	})

	t.Run("basic fields", func(t *testing.T) {
		incidents := []Incident{{
			ID:        "inc-1",
			Name:      "DB outage",
			Status:    IncidentStatusRef{Name: "Active"},
			CreatedAt: "2026-01-01T00:00:00Z",
		}}
		result := ToCompact(incidents)
		if len(result) != 1 {
			t.Fatalf("expected 1 item, got %d", len(result))
		}
		c := result[0]
		if c.ID != "inc-1" || c.Name != "DB outage" || c.Status != "Active" || c.CreatedAt != "2026-01-01T00:00:00Z" {
			t.Errorf("unexpected compact: %+v", c)
		}
	})

	t.Run("nil severity omitted", func(t *testing.T) {
		incidents := []Incident{{
			ID:     "inc-1",
			Name:   "test",
			Status: IncidentStatusRef{Name: "Active"},
		}}
		result := ToCompact(incidents)
		if result[0].Severity != "" {
			t.Errorf("expected empty severity, got %q", result[0].Severity)
		}
	})

	t.Run("severity populated", func(t *testing.T) {
		incidents := []Incident{{
			ID:       "inc-1",
			Name:     "test",
			Status:   IncidentStatusRef{Name: "Active"},
			Severity: &IncidentSeverity{ID: "sev-1", Name: "Critical"},
		}}
		result := ToCompact(incidents)
		if result[0].Severity != "Critical" {
			t.Errorf("expected Critical, got %q", result[0].Severity)
		}
	})

	t.Run("incident lead from role assignments", func(t *testing.T) {
		incidents := []Incident{{
			ID:     "inc-1",
			Name:   "test",
			Status: IncidentStatusRef{Name: "Active"},
			IncidentRoleAssignments: []IncidentRoleAssignment{
				{
					Role:     &IncidentRole{ID: "role-1", Name: "Reporter"},
					Assignee: &IncidentUser{ID: "u-1", Name: "Alice"},
				},
				{
					Role:     &IncidentRole{ID: "role-2", Name: "Incident Lead"},
					Assignee: &IncidentUser{ID: "u-2", Name: "Bob"},
				},
			},
		}}
		result := ToCompact(incidents)
		if result[0].IncidentLead != "Bob" {
			t.Errorf("expected Bob, got %q", result[0].IncidentLead)
		}
	})

	t.Run("no matching role", func(t *testing.T) {
		incidents := []Incident{{
			ID:     "inc-1",
			Name:   "test",
			Status: IncidentStatusRef{Name: "Active"},
			IncidentRoleAssignments: []IncidentRoleAssignment{
				{
					Role:     &IncidentRole{ID: "role-1", Name: "Reporter"},
					Assignee: &IncidentUser{ID: "u-1", Name: "Alice"},
				},
			},
		}}
		result := ToCompact(incidents)
		if result[0].IncidentLead != "" {
			t.Errorf("expected empty incident lead, got %q", result[0].IncidentLead)
		}
	})

	t.Run("no role assignments", func(t *testing.T) {
		incidents := []Incident{{
			ID:     "inc-1",
			Name:   "test",
			Status: IncidentStatusRef{Name: "Active"},
		}}
		result := ToCompact(incidents)
		if result[0].IncidentLead != "" {
			t.Errorf("expected empty incident lead, got %q", result[0].IncidentLead)
		}
	})

	t.Run("name truncated at 200 chars", func(t *testing.T) {
		longName := strings.Repeat("x", 250)
		incidents := []Incident{{
			ID:     "inc-1",
			Name:   longName,
			Status: IncidentStatusRef{Name: "Active"},
		}}
		result := ToCompact(incidents)
		if len(result[0].Name) != 200 {
			t.Errorf("expected name length 200, got %d", len(result[0].Name))
		}
	})
}
