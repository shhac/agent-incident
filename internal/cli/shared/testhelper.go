package shared

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/api"
)

// NewTestRoot creates a root cobra.Command and registers a domain via the
// provided register function. Used by domain test packages to avoid
// duplicating test scaffolding.
func NewTestRoot(register func(*cobra.Command, GlobalsFunc)) *cobra.Command {
	root := &cobra.Command{
		Use:           "agent-incident",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	globals := func() *GlobalFlags {
		return &GlobalFlags{}
	}
	register(root, globals)
	return root
}

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
