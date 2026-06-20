package cli

import (
	"strings"
	"testing"
)

// A missing required flag must surface as an error from Execute so the root Run
// sink renders it once and exits non-zero — not a swallowed nil (exit 0).
func TestMissingRequiredFlagReturnsError(t *testing.T) {
	root := newRootCmd("test")
	root.SetArgs([]string{"incident", "create"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing --name, got nil")
	}
	if !strings.Contains(err.Error(), "--name is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// An unknown subcommand under a domain group must return a structured error
// (exit 1), mirroring the root's own unknown-command handling.
func TestGroupUnknownSubcommandReturnsError(t *testing.T) {
	for _, args := range [][]string{
		{"incident", "bogus"},
		{"oncall", "bogus"},
		{"oncall", "schedule", "bogus"},
		{"ref", "catalog", "bogus"},
		{"oncall", "escalation", "path", "bogus"},
	} {
		root := newRootCmd("test")
		root.SetArgs(args)

		err := root.Execute()
		if err == nil {
			t.Fatalf("%v: expected error for unknown subcommand, got nil", args)
		}
		if !strings.Contains(err.Error(), "unknown command") {
			t.Fatalf("%v: unexpected error: %v", args, err)
		}
	}
}

// A group invoked with no args still falls through to help (exit 0), unchanged
// by the unknown-command handler.
func TestGroupNoArgsIsNotAnError(t *testing.T) {
	root := newRootCmd("test")
	root.SetArgs([]string{"incident"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected nil for group help, got: %v", err)
	}
}
