package shared

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shhac/agent-incident/internal/api"
)

// SetupMockServer creates an httptest.Server and injects it via ClientFactory.
// The server and factory are cleaned up when the test completes.
func SetupMockServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(func() {
		srv.Close()
		ClientFactory = nil
	})
	ClientFactory = func() (*api.Client, error) {
		return api.NewTestClient(srv.URL, "test-api-key"), nil
	}
	return srv
}
