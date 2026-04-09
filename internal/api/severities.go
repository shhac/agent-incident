package api

import (
	"context"
	"fmt"
	"net/http"
)

type Severity struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Rank        int    `json:"rank"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type severitiesWrapper struct {
	Severities []Severity `json:"severities"`
}

type severityWrapper struct {
	Severity Severity `json:"severity"`
}

func (c *Client) ListSeverities(ctx context.Context) ([]Severity, error) {
	result, err := doAndDecode[severitiesWrapper](c, ctx, http.MethodGet, "/v1/severities", nil)
	if err != nil {
		return nil, err
	}
	return result.Severities, nil
}

func (c *Client) GetSeverity(ctx context.Context, id string) (*Severity, error) {
	return doAndDecodeField[severityWrapper, Severity](c, ctx, http.MethodGet, fmt.Sprintf("/v1/severities/%s", id), nil,
		func(w *severityWrapper) *Severity { return &w.Severity })
}
