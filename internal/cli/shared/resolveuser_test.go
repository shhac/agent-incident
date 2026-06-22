package shared

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shhac/agent-incident/internal/api"
)

// usersServer returns an httptest server that serves the given users from
// /v2/users, mirroring the incident.io list-users envelope. It is the small
// ListUsers stub used to drive ResolveUserID through its match ladder.
func usersServer(t *testing.T, users []api.User) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"users":           users,
			"pagination_meta": map[string]any{},
		})
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestResolveUserID_MatchLadder(t *testing.T) {
	users := []api.User{
		{ID: "u-alice", Name: "Alice Smith", Email: "alice@example.com"},
		{ID: "u-bob", Name: "Bob Jones", Email: "bob@example.com"},
		{ID: "u-alicia", Name: "Alicia Stone", Email: "alicia@example.com"},
		// Two users sharing the same name to exercise the ambiguous-exact branch.
		{ID: "u-dup1", Name: "Sam Twin", Email: "sam1@example.com"},
		{ID: "u-dup2", Name: "Sam Twin", Email: "sam2@example.com"},
	}

	tests := []struct {
		name      string
		ref       string
		wantID    string
		wantErr   bool
		errSubstr string
	}{
		{
			name:   "exact email match",
			ref:    "alice@example.com",
			wantID: "u-alice",
		},
		{
			name:   "exact name match",
			ref:    "Bob Jones",
			wantID: "u-bob",
		},
		{
			name:      "ambiguous exact",
			ref:       "Sam Twin",
			wantErr:   true,
			errSubstr: "ambiguous",
		},
		{
			name:      "ambiguous substring",
			ref:       "Alic",
			wantErr:   true,
			errSubstr: "ambiguous",
		},
		{
			name:   "single substring",
			ref:    "Jones",
			wantID: "u-bob",
		},
		{
			name:      "not found",
			ref:       "nobody",
			wantErr:   true,
			errSubstr: "no user found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := usersServer(t, users)
			client := api.NewTestClient(srv.URL, "test-key")

			id, err := ResolveUserID(context.Background(), client, tt.ref)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got id %q", id)
				}
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != tt.wantID {
				t.Errorf("ResolveUserID(%q) = %q, want %q", tt.ref, id, tt.wantID)
			}
		})
	}
}
