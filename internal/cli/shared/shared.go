package shared

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/config"
	"github.com/shhac/agent-incident/internal/credential"
	agenterrors "github.com/shhac/agent-incident/internal/errors"
	"github.com/shhac/agent-incident/internal/output"
)

// GlobalFlags holds persistent flags available to all commands.
type GlobalFlags struct {
	Org     string
	APIKey  string
	Format  string
	Timeout int
	Debug   bool
}

// GlobalsFunc is the signature for the globals accessor passed to domain Register functions.
type GlobalsFunc = func() *GlobalFlags

func MakeContext(timeoutMs int) (context.Context, context.CancelFunc) {
	if timeoutMs > 0 {
		return context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
	}
	return context.WithCancel(context.Background())
}

// ResolveOrg determines which organization alias to use.
// Resolution: --organization flag > INCIDENT_API_KEY env shortcircuit > config default.
func ResolveOrg(orgAlias string) (string, error) {
	if orgAlias != "" {
		return orgAlias, nil
	}
	if os.Getenv("INCIDENT_API_KEY") != "" {
		return "", nil // env key will be used directly
	}
	cfg := config.Read()
	if cfg.DefaultOrg != "" {
		return cfg.DefaultOrg, nil
	}
	available := make([]string, 0)
	for name := range cfg.Organizations {
		available = append(available, name)
	}
	hint := "No organizations configured. Add one with 'agent-incident auth add <alias> --api-key <key>'"
	if len(available) > 0 {
		hint = fmt.Sprintf("Available organizations: %s. Set a default with 'agent-incident auth default <alias>'", strings.Join(available, ", "))
	}
	return "", agenterrors.New("no organization specified", agenterrors.FixableByAgent).WithHint(hint)
}

// NewClientFromFlags resolves credentials and returns an API client.
// Resolution: --api-key flag > INCIDENT_API_KEY env > credential store lookup.
func NewClientFromFlags(apiKeyFlag, orgAlias string) (*api.Client, error) {
	if apiURL := os.Getenv("INCIDENT_API_URL"); apiURL != "" {
		apiKey := apiKeyFlag
		if apiKey == "" {
			apiKey = os.Getenv("INCIDENT_API_KEY")
		}
		return api.NewTestClient(apiURL, apiKey), nil
	}

	if apiKeyFlag != "" {
		return api.NewClient(apiKeyFlag), nil
	}

	if envKey := os.Getenv("INCIDENT_API_KEY"); envKey != "" {
		return api.NewClient(envKey), nil
	}

	alias, err := ResolveOrg(orgAlias)
	if err != nil {
		return nil, err
	}

	cred, err := credential.Get(alias)
	if err != nil {
		var nf *credential.NotFoundError
		if errors.As(err, &nf) {
			return nil, agenterrors.Newf(agenterrors.FixableByHuman, "credentials for organization %q not found", alias).
				WithHint("Add credentials with 'agent-incident auth add " + alias + " --api-key <key>'")
		}
		return nil, agenterrors.Wrap(err, agenterrors.FixableByHuman)
	}

	if cred.APIKey == "" {
		return nil, agenterrors.Newf(agenterrors.FixableByHuman, "organization %q has no API key", alias).
			WithHint("Update with 'agent-incident auth add " + alias + " --api-key <key>'")
	}

	return api.NewClient(cred.APIKey), nil
}

// ClientFactory allows DI for testing. When set, WithClient uses this instead of real credential resolution.
var ClientFactory func() (*api.Client, error)

// WithClient resolves an API client and runs the callback. Errors are written to stderr.
func WithClient(g *GlobalFlags, fn func(ctx context.Context, client *api.Client) error) error {
	ctx, cancel := MakeContext(g.Timeout)
	defer cancel()

	var client *api.Client
	var err error
	if ClientFactory != nil {
		client, err = ClientFactory()
	} else {
		client, err = NewClientFromFlags(g.APIKey, g.Org)
	}
	if err != nil {
		output.WriteError(os.Stderr, err)
		return nil
	}

	client.SetDebug(g.Debug)

	if err := fn(ctx, client); err != nil {
		output.WriteError(os.Stderr, err)
	}
	return nil
}

func ToAnySlice[T any](s []T) []any {
	result := make([]any, len(s))
	for i, v := range s {
		result[i] = v
	}
	return result
}

// WritePaginatedList writes a list in the resolved format (default: NDJSON).
func WritePaginatedList(items []any, pagination *output.Pagination, format string) {
	f := output.ResolveFormat(format, output.FormatNDJSON)
	if f == output.FormatNDJSON {
		w := output.NewNDJSONWriter(os.Stdout)
		for _, item := range items {
			_ = w.WriteItem(item)
		}
		if pagination != nil {
			_ = w.WritePagination(pagination)
		}
		return
	}
	result := map[string]any{"data": items}
	if pagination != nil {
		result["pagination"] = pagination
	}
	output.Print(result, f, true)
}

// WriteItem writes a single item in the resolved format (default: JSON).
func WriteItem(data any, format string) {
	f := output.ResolveFormat(format, output.FormatJSON)
	output.Print(data, f, true)
}

// CursorPagination builds pagination metadata from a cursor string.
func CursorPagination(cursor string) *output.Pagination {
	if cursor == "" {
		return nil
	}
	return &output.Pagination{HasMore: true, NextCursor: cursor}
}

// RequireFlag checks that a flag value is non-empty, writing an error to stderr if not.
func RequireFlag(flag, value, hint string) bool {
	if value != "" {
		return true
	}
	err := agenterrors.Newf(agenterrors.FixableByAgent, "--%s is required", flag)
	if hint != "" {
		err = err.WithHint(hint)
	}
	output.WriteError(os.Stderr, err)
	return false
}
