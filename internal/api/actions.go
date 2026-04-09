package api

import (
	"context"
	"net/http"
	"net/url"
)

type Action struct {
	ID          string `json:"id"`
	IncidentID  string `json:"incident_id"`
	Status      string `json:"status"`
	Description string `json:"description,omitempty"`
	Assignee    *User  `json:"assignee,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	CompletedAt string `json:"completed_at,omitempty"`
}

type actionsWrapper struct {
	Actions        []Action        `json:"actions"`
	PaginationMeta *paginationMeta `json:"pagination_meta,omitempty"`
}

type actionWrapper struct {
	Action Action `json:"action"`
}

func (c *Client) ListActions(ctx context.Context, incidentID string, pageSize int, after string) ([]Action, string, error) {
	params := url.Values{}
	if incidentID != "" {
		params.Set("incident_id", incidentID)
	}
	addPaginationParams(params, pageSize, after)
	result, err := doAndDecode[actionsWrapper](c, ctx, http.MethodGet, buildPath("/v2/actions", params), nil)
	if err != nil {
		return nil, "", err
	}
	return result.Actions, extractCursor(result.PaginationMeta), nil
}

func (c *Client) GetAction(ctx context.Context, id string) (*Action, error) {
	return doAndDecodeField[actionWrapper, Action](c, ctx, http.MethodGet, "/v2/actions/"+id, nil,
		func(w *actionWrapper) *Action { return &w.Action })
}
