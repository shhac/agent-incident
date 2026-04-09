package api

import (
	"testing"
)

func TestToCompactUsers(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		result := ToCompactUsers(nil)
		if len(result) != 0 {
			t.Fatalf("expected empty slice, got %d items", len(result))
		}
	})

	t.Run("single user drops extra fields", func(t *testing.T) {
		users := []User{{
			ID:          "user-1",
			Name:        "Alice Smith",
			Email:       "alice@example.com",
			Role:        "admin",
			SlackUserID: "U12345",
			CreatedAt:   "2024-01-01T00:00:00Z",
			UpdatedAt:   "2024-06-15T12:00:00Z",
		}}
		result := ToCompactUsers(users)
		if len(result) != 1 {
			t.Fatalf("expected 1 item, got %d", len(result))
		}
		c := result[0]
		if c.ID != "user-1" {
			t.Errorf("ID = %q, want user-1", c.ID)
		}
		if c.Name != "Alice Smith" {
			t.Errorf("Name = %q, want Alice Smith", c.Name)
		}
		if c.Email != "alice@example.com" {
			t.Errorf("Email = %q, want alice@example.com", c.Email)
		}
		if c.Role != "admin" {
			t.Errorf("Role = %q, want admin", c.Role)
		}
	})

	t.Run("multiple users preserve order", func(t *testing.T) {
		users := []User{
			{ID: "user-1", Name: "Alice", Email: "alice@example.com", Role: "admin", SlackUserID: "U1", CreatedAt: "2024-01-01T00:00:00Z", UpdatedAt: "2024-01-01T00:00:00Z"},
			{ID: "user-2", Name: "Bob", Email: "bob@example.com", Role: "responder", SlackUserID: "U2", CreatedAt: "2024-02-01T00:00:00Z", UpdatedAt: "2024-02-01T00:00:00Z"},
			{ID: "user-3", Name: "Carol", Email: "carol@example.com", Role: "viewer", SlackUserID: "U3", CreatedAt: "2024-03-01T00:00:00Z", UpdatedAt: "2024-03-01T00:00:00Z"},
		}
		result := ToCompactUsers(users)
		if len(result) != 3 {
			t.Fatalf("expected 3 items, got %d", len(result))
		}
		if result[0].ID != "user-1" || result[1].ID != "user-2" || result[2].ID != "user-3" {
			t.Errorf("order not preserved: got %q, %q, %q", result[0].ID, result[1].ID, result[2].ID)
		}
		if result[1].Name != "Bob" || result[1].Email != "bob@example.com" || result[1].Role != "responder" {
			t.Errorf("field mapping wrong for user-2: %+v", result[1])
		}
	})
}
