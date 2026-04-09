package api

import (
	"context"
	"net/http"
)

// Identity represents the response from GET /v1/identity.
type Identity struct {
	Name  string        `json:"name"`
	Roles []string      `json:"roles"`
	Scope []string      `json:"scope,omitempty"`
	Actor *IdentityActor `json:"actor,omitempty"`
}

type IdentityActor struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type identityWrapper struct {
	Identity Identity `json:"identity"`
}

func (c *Client) GetIdentity(ctx context.Context) (*Identity, error) {
	return doAndDecodeField[identityWrapper, Identity](c, ctx, http.MethodGet, "/v1/identity", nil,
		func(w *identityWrapper) *Identity { return &w.Identity })
}
