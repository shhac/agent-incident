package api

import (
	"testing"
	"time"
)

func TestToCompactAlerts(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		result := ToCompactAlerts(nil)
		if len(result) != 0 {
			t.Fatalf("expected empty slice, got %d items", len(result))
		}
	})

	t.Run("single alert", func(t *testing.T) {
		now := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
		alerts := []Alert{{
			ID:          "alert-1",
			Title:       "CPU spike",
			Description: "CPU usage exceeded 90%",
			Status:      "firing",
			SourceURL:   "https://example.com",
			CreatedAt:   now,
			UpdatedAt:   now,
		}}
		result := ToCompactAlerts(alerts)
		if len(result) != 1 {
			t.Fatalf("expected 1 item, got %d", len(result))
		}
		c := result[0]
		if c.ID != "alert-1" {
			t.Errorf("ID = %q, want alert-1", c.ID)
		}
		if c.Title != "CPU spike" {
			t.Errorf("Title = %q, want CPU spike", c.Title)
		}
		if c.Status != "firing" {
			t.Errorf("Status = %q, want firing", c.Status)
		}
		if !c.CreatedAt.Equal(now) {
			t.Errorf("CreatedAt = %v, want %v", c.CreatedAt, now)
		}
	})

	t.Run("multiple alerts preserve order", func(t *testing.T) {
		t1 := time.Date(2026, 4, 9, 10, 0, 0, 0, time.UTC)
		t2 := time.Date(2026, 4, 9, 11, 0, 0, 0, time.UTC)
		alerts := []Alert{
			{ID: "a-1", Title: "First", Status: "firing", CreatedAt: t1, UpdatedAt: t1},
			{ID: "a-2", Title: "Second", Status: "resolved", CreatedAt: t2, UpdatedAt: t2},
		}
		result := ToCompactAlerts(alerts)
		if len(result) != 2 {
			t.Fatalf("expected 2 items, got %d", len(result))
		}
		if result[0].ID != "a-1" || result[1].ID != "a-2" {
			t.Errorf("order not preserved: got %q, %q", result[0].ID, result[1].ID)
		}
	})

	t.Run("description and source_url not carried over", func(t *testing.T) {
		now := time.Now()
		alerts := []Alert{{
			ID:          "a-1",
			Title:       "test",
			Description: "should be dropped",
			Status:      "firing",
			SourceURL:   "https://example.com",
			CreatedAt:   now,
			UpdatedAt:   now,
		}}
		result := ToCompactAlerts(alerts)
		// AlertCompact has no Description or SourceURL fields,
		// so this test verifies the struct only carries the expected fields.
		if result[0].ID != "a-1" || result[0].Title != "test" || result[0].Status != "firing" {
			t.Errorf("unexpected compact: %+v", result[0])
		}
	})
}
