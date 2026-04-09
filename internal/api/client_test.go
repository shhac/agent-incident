package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	agenterrors "github.com/shhac/agent-incident/internal/errors"
)

func TestSuccessfulGETWithAuthHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-key-123" {
			t.Errorf("expected Authorization %q, got %q", "Bearer test-key-123", auth)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer srv.Close()

	client := NewTestClient(srv.URL, "test-key-123")
	raw, err := client.do(context.Background(), http.MethodGet, "/v1/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]string
	if e := json.Unmarshal(raw, &result); e != nil {
		t.Fatalf("invalid JSON: %v", e)
	}
	if result["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", result["status"])
	}
}

func TestClassifyHTTPError(t *testing.T) {
	tests := []struct {
		status    int
		body      string
		wantFix   agenterrors.FixableBy
		wantHint  bool
		wantInMsg string
	}{
		{401, `{"error":"invalid key"}`, agenterrors.FixableByHuman, true, "Authentication failed"},
		{403, `{"error":"forbidden"}`, agenterrors.FixableByHuman, true, "Permission denied"},
		{404, `{"error":"not found"}`, agenterrors.FixableByAgent, true, "Not found"},
		{429, `{}`, agenterrors.FixableByRetry, true, "Rate limited"},
		{500, `{"message":"internal error"}`, agenterrors.FixableByRetry, true, "incident.io API error"},
		{502, `{"error":"bad gateway"}`, agenterrors.FixableByRetry, true, "incident.io API error"},
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			apiErr := classifyHTTPError(tt.status, []byte(tt.body))
			if apiErr.FixableBy != tt.wantFix {
				t.Fatalf("expected fixable_by %q, got %q", tt.wantFix, apiErr.FixableBy)
			}
			if tt.wantHint && apiErr.Hint == "" {
				t.Fatal("expected hint to be set")
			}
			if tt.wantInMsg != "" && !contains(apiErr.Message, tt.wantInMsg) {
				t.Fatalf("expected message to contain %q, got %q", tt.wantInMsg, apiErr.Message)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestBuildPath(t *testing.T) {
	t.Run("no params", func(t *testing.T) {
		got := buildPath("/v1/incidents", url.Values{})
		if got != "/v1/incidents" {
			t.Fatalf("expected %q, got %q", "/v1/incidents", got)
		}
	})

	t.Run("with params", func(t *testing.T) {
		params := url.Values{}
		params.Set("page_size", "25")
		params.Set("status", "active")
		got := buildPath("/v1/incidents", params)
		if got == "/v1/incidents" {
			t.Fatal("expected query string to be appended")
		}
		parsed, err := url.Parse(got)
		if err != nil {
			t.Fatalf("invalid URL: %v", err)
		}
		if parsed.Query().Get("page_size") != "25" {
			t.Fatal("expected page_size=25")
		}
		if parsed.Query().Get("status") != "active" {
			t.Fatal("expected status=active")
		}
	})
}

func TestDoAndDecode(t *testing.T) {
	t.Run("valid JSON", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"name": "test-incident"})
		}))
		defer srv.Close()

		client := NewTestClient(srv.URL, "key")
		type Result struct {
			Name string `json:"name"`
		}
		result, err := doAndDecode[Result](client, context.Background(), http.MethodGet, "/test", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Name != "test-incident" {
			t.Fatalf("expected %q, got %q", "test-incident", result.Name)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("not json at all"))
		}))
		defer srv.Close()

		client := NewTestClient(srv.URL, "key")
		type Result struct {
			Name string `json:"name"`
		}
		_, err := doAndDecode[Result](client, context.Background(), http.MethodGet, "/test", nil)
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
		var apiErr *agenterrors.APIError
		if !agenterrors.As(err, &apiErr) {
			t.Fatalf("expected APIError, got %T", err)
		}
	})
}

func TestHTTPErrorStatusCodes(t *testing.T) {
	statuses := []int{401, 403, 404, 429, 500}
	for _, status := range statuses {
		t.Run(http.StatusText(status), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
				w.Write([]byte(`{"error":"test error"}`))
			}))
			defer srv.Close()

			client := NewTestClient(srv.URL, "key")
			_, err := client.do(context.Background(), http.MethodGet, "/test", nil)
			if err == nil {
				t.Fatalf("expected error for status %d", status)
			}
			var apiErr *agenterrors.APIError
			if !agenterrors.As(err, &apiErr) {
				t.Fatalf("expected APIError for status %d, got %T", status, err)
			}
		})
	}
}

func TestExtractErrorMessage(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{
			name:   "JSON errors array",
			status: 400,
			body:   `{"errors": ["bad request"]}`,
			want:   "bad request",
		},
		{
			name:   "JSON error field",
			status: 404,
			body:   `{"error": "not found"}`,
			want:   "not found",
		},
		{
			name:   "JSON message field",
			status: 500,
			body:   `{"message": "internal error"}`,
			want:   "internal error",
		},
		{
			name:   "non-JSON body short",
			status: 418,
			body:   "something went wrong",
			want:   "HTTP 418: something went wrong",
		},
		{
			name:   "non-JSON body exceeds 200 bytes",
			status: 400,
			body:   string(make([]byte, 201)),
			want:   "HTTP 400",
		},
		{
			name:   "empty body",
			status: 502,
			body:   "",
			want:   "HTTP 502",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractErrorMessage(tt.status, []byte(tt.body))
			if got != tt.want {
				t.Errorf("extractErrorMessage(%d, %q) = %q, want %q", tt.status, tt.body, got, tt.want)
			}
		})
	}
}

func TestExtractErrorMessageViaHTTP(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		body      string
		wantInMsg string
	}{
		{
			name:      "JSON errors array via do",
			status:    422,
			body:      `{"errors": ["bad request"]}`,
			wantInMsg: "bad request",
		},
		{
			name:      "JSON error field via do",
			status:    404,
			body:      `{"error": "not found"}`,
			wantInMsg: "not found",
		},
		{
			name:      "JSON message field via do",
			status:    500,
			body:      `{"message": "internal error"}`,
			wantInMsg: "internal error",
		},
		{
			name:      "non-JSON short body via do",
			status:    418,
			body:      "something went wrong",
			wantInMsg: "HTTP 418: something went wrong",
		},
		{
			name:      "empty body via do",
			status:    502,
			body:      "",
			wantInMsg: "HTTP 502",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			client := NewTestClient(srv.URL, "key")
			_, err := client.do(context.Background(), http.MethodGet, "/test", nil)
			if err == nil {
				t.Fatal("expected error")
			}
			var apiErr *agenterrors.APIError
			if !agenterrors.As(err, &apiErr) {
				t.Fatalf("expected APIError, got %T", err)
			}
			if !contains(apiErr.Message, tt.wantInMsg) {
				t.Errorf("expected message containing %q, got %q", tt.wantInMsg, apiErr.Message)
			}
		})
	}
}

func TestPOSTWithBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		ct := r.Header.Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", ct)
		}
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "new-incident" {
			t.Errorf("expected body name %q, got %q", "new-incident", body["name"])
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": "inc-123"})
	}))
	defer srv.Close()

	client := NewTestClient(srv.URL, "key")
	raw, err := client.do(context.Background(), http.MethodPost, "/v1/incidents", map[string]string{"name": "new-incident"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]string
	json.Unmarshal(raw, &result)
	if result["id"] != "inc-123" {
		t.Fatalf("expected id %q, got %q", "inc-123", result["id"])
	}
}
