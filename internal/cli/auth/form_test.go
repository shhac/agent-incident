package auth

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/shhac/agent-incident/internal/config"
	"github.com/shhac/agent-incident/internal/credential"
	agenterrors "github.com/shhac/agent-incident/internal/errors"
	"github.com/shhac/lib-agent-cli/dialog"
	"github.com/shhac/lib-agent-cli/dialog/dialogtest"
)

func TestAuthAddFormFillsMissingKeyEndToEnd(t *testing.T) {
	config.SetConfigDir(t.TempDir())

	rec := &dialogtest.Recorder{
		PromptResults: []dialog.Result{{ID: "secret", Value: "dialog-key-456"}},
	}
	defer dialog.SetDefault(rec)()

	root := newTestRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"auth", "add", "formorg", "--form"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rec.Calls) != 1 {
		t.Fatalf("expected the dialog to be prompted once, got %d calls", len(rec.Calls))
	}

	cred, err := credential.Get("formorg")
	if err != nil {
		t.Fatalf("credential not found after add: %v", err)
	}
	if cred.APIKey != "dialog-key-456" && !cred.KeychainManaged {
		t.Errorf("expected api key from dialog or keychain managed, got key=%q keychain=%v", cred.APIKey, cred.KeychainManaged)
	}
}

func TestPromptSecretReturnsEarlyWhenFlagSupplied(t *testing.T) {
	rec := &dialogtest.Recorder{
		PromptResults: []dialog.Result{{ID: "secret", Value: "should not be used"}},
	}
	defer dialog.SetDefault(rec)()

	key, err := promptSecretViaDialog(context.Background(), "acme", "flag-key")
	if err != nil {
		t.Fatalf("promptSecretViaDialog() error = %v", err)
	}
	if key != "flag-key" {
		t.Fatalf("returned key = %q, want flag-key", key)
	}
	if len(rec.Calls) != 0 {
		t.Errorf("Prompt should not have been called, got %d calls", len(rec.Calls))
	}
}

func TestPromptSecretFillsMissingKeyFromDialog(t *testing.T) {
	rec := &dialogtest.Recorder{
		PromptResults: []dialog.Result{{ID: "secret", Value: "from-dialog"}},
	}
	defer dialog.SetDefault(rec)()

	key, err := promptSecretViaDialog(context.Background(), "acme", "")
	if err != nil {
		t.Fatalf("promptSecretViaDialog() error = %v", err)
	}
	if key != "from-dialog" {
		t.Errorf("key = %q, want 'from-dialog'", key)
	}
	if len(rec.Calls) != 1 {
		t.Fatalf("expected 1 prompt call, got %d", len(rec.Calls))
	}
	spec := rec.Calls[0]
	if len(spec.Items) != 1 {
		t.Fatalf("expected 1 field in spec, got %d", len(spec.Items))
	}
	if spec.Items[0].InputType != dialog.Password {
		t.Errorf("field InputType = %v, want Password", spec.Items[0].InputType)
	}
	if !strings.Contains(spec.Title, "acme") {
		t.Errorf("spec title = %q, want it to contain the alias", spec.Title)
	}
}

func TestPromptSecretReturnsHumanErrorWhenNoGUI(t *testing.T) {
	rec := &dialogtest.Recorder{
		AvailableErr: fmt.Errorf("%w: SSH session detected", dialog.ErrNoGUI),
	}
	defer dialog.SetDefault(rec)()

	_, err := promptSecretViaDialog(context.Background(), "acme", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var aerr *agenterrors.APIError
	if !errors.As(err, &aerr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if aerr.FixableBy != agenterrors.FixableByHuman {
		t.Errorf("FixableBy = %q, want human", aerr.FixableBy)
	}
	if !strings.Contains(aerr.Hint, "graphical desktop") {
		t.Errorf("hint = %q, want it to mention graphical desktop fallback", aerr.Hint)
	}
	if !strings.Contains(aerr.Hint, "--api-key") {
		t.Errorf("hint = %q, want it to suggest the non-interactive fallback", aerr.Hint)
	}
	// Sentinel chain must be preserved so callers can errors.Is downstream.
	if !errors.Is(err, dialog.ErrNoGUI) {
		t.Errorf("errors.Is(err, ErrNoGUI) = false, want true (sentinel chain broken)")
	}
}

func TestPromptSecretReturnsRetryErrorOnCancel(t *testing.T) {
	rec := &dialogtest.Recorder{
		PromptErr: fmt.Errorf("%w (incident.io API key)", dialog.ErrCancelled),
	}
	defer dialog.SetDefault(rec)()

	_, err := promptSecretViaDialog(context.Background(), "acme", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var aerr *agenterrors.APIError
	if !errors.As(err, &aerr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if aerr.FixableBy != agenterrors.FixableByRetry {
		t.Errorf("FixableBy = %q, want retry", aerr.FixableBy)
	}
	if !strings.Contains(aerr.Hint, "cancelled") && !strings.Contains(aerr.Hint, "Re-run") {
		t.Errorf("hint = %q, should mention cancellation and re-run", aerr.Hint)
	}
	// Sentinel chain must be preserved so callers can errors.Is downstream.
	if !errors.Is(err, dialog.ErrCancelled) {
		t.Errorf("errors.Is(err, ErrCancelled) = false, want true (sentinel chain broken)")
	}
}
