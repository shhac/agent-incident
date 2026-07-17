package auth

import (
	"context"

	agenterrors "github.com/shhac/agent-incident/internal/errors"
	"github.com/shhac/lib-agent-cli/dialog"
)

// promptSecretViaDialog asks the user (via a native OS dialog) for the API key
// when it wasn't supplied on the command line. The LLM driving the CLI never
// sees what the user types — input goes directly into the OS dialog, and only a
// redacted receipt is returned on stdout. When apiKey is already set, it is
// returned unchanged and no dialog is shown.
func promptSecretViaDialog(ctx context.Context, alias, apiKey string) (string, error) {
	if apiKey != "" {
		return apiKey, nil
	}

	if err := dialog.Available(); err != nil {
		return apiKey, classifyDialogErr(err, alias)
	}

	value, err := dialog.PromptSecret(ctx, "agent-incident auth: "+alias, "incident.io API key")
	if err != nil {
		return apiKey, classifyDialogErr(err, alias)
	}
	return value, nil
}

// classifyDialogErr adapts a dialog error to agent-incident's output contract.
// dialog.ClassifyError owns the sentinel→category mapping; this only augments
// the hint with agent-incident-specific guidance. The sentinel chain is
// preserved so callers can errors.Is downstream.
func classifyDialogErr(err error, alias string) error {
	category, baseHint := dialog.ClassifyError(err)
	hint := baseHint
	var fixableBy agenterrors.FixableBy
	switch category {
	case dialog.CategoryHuman:
		fixableBy = agenterrors.FixableByHuman
		hint = "agent-incident auth add --form requires a graphical desktop session. " +
			"Ask the user to run on their local machine, or fall back to non-interactive: " +
			"agent-incident auth add " + alias + " --api-key <key>"
	case dialog.CategoryRetry:
		fixableBy = agenterrors.FixableByRetry
		hint = "User cancelled the dialog. Re-run agent-incident auth add --form to retry."
	default:
		fixableBy = agenterrors.FixableByAgent
	}
	return agenterrors.Wrap(err, fixableBy).WithHint(hint)
}
