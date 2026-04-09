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
		return err
	}

	client.SetDebug(g.Debug)

	if err := fn(ctx, client); err != nil {
		output.WriteError(os.Stderr, err)
		return err
	}
	return nil
}
