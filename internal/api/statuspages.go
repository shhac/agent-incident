package api

import (
	"context"
	"net/http"
	"net/url"
)

type StatusPage struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	URL       string `json:"public_url,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type StatusPageIncident struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type StatusPageIncidentUpdate struct {
	ID        string `json:"id"`
	Message   string `json:"message,omitempty"`
	ToStatus  string `json:"to_status,omitempty"`
	CreatedAt string `json:"created_at"`
}

type statusPagesWrapper struct {
	StatusPages []StatusPage `json:"status_pages"`
}

type statusPageIncidentsWrapper struct {
	StatusPageIncidents []StatusPageIncident `json:"status_page_incidents"`
	PaginationMeta      *paginationMeta      `json:"pagination_meta,omitempty"`
}

type statusPageIncidentWrapper struct {
	StatusPageIncident StatusPageIncident `json:"status_page_incident"`
}

type CreateStatusPageIncidentParams struct {
	StatusPageID string `json:"status_page_id"`
	Name         string `json:"name"`
}

type UpdateStatusPageIncidentParams struct {
	Status string `json:"status,omitempty"`
}

type CreateStatusPageIncidentUpdateParams struct {
	StatusPageIncidentID string `json:"status_page_incident_id"`
	Message              string `json:"message,omitempty"`
	ToStatus             string `json:"to_status,omitempty"`
}

func (c *Client) ListStatusPages(ctx context.Context) ([]StatusPage, error) {
	result, err := doAndDecode[statusPagesWrapper](c, ctx, http.MethodGet, "/v2/status_pages", nil)
	if err != nil {
		return nil, err
	}
	return result.StatusPages, nil
}

func (c *Client) ListStatusPageIncidents(ctx context.Context, pageID string) ([]StatusPageIncident, error) {
	params := url.Values{}
	if pageID != "" {
		params.Set("status_page_id", pageID)
	}
	result, err := doAndDecode[statusPageIncidentsWrapper](c, ctx, http.MethodGet, buildPath("/v2/status_page_incidents", params), nil)
	if err != nil {
		return nil, err
	}
	return result.StatusPageIncidents, nil
}

func (c *Client) CreateStatusPageIncident(ctx context.Context, params CreateStatusPageIncidentParams) (*StatusPageIncident, error) {
	return doAndDecodeField[statusPageIncidentWrapper, StatusPageIncident](
		c, ctx, http.MethodPost, "/v2/status_page_incidents", params,
		func(w *statusPageIncidentWrapper) *StatusPageIncident { return &w.StatusPageIncident })
}

func (c *Client) UpdateStatusPageIncident(ctx context.Context, id string, params UpdateStatusPageIncidentParams) (*StatusPageIncident, error) {
	return doAndDecodeField[statusPageIncidentWrapper, StatusPageIncident](
		c, ctx, http.MethodPut, "/v2/status_page_incidents/"+id, params,
		func(w *statusPageIncidentWrapper) *StatusPageIncident { return &w.StatusPageIncident })
}

func (c *Client) CreateStatusPageIncidentUpdate(ctx context.Context, params CreateStatusPageIncidentUpdateParams) error {
	_, err := c.do(ctx, http.MethodPost, "/v2/status_page_incident_updates", params)
	return err
}
