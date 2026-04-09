package api

import (
	"context"
	"fmt"
	"net/http"
)

// IncidentStatusResource is the full status object from the incident_statuses API.
// This is distinct from IncidentStatus in incidents.go which is the embedded summary.
type IncidentStatusResource struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Rank        int    `json:"rank"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type incidentStatusesWrapper struct {
	IncidentStatuses []IncidentStatusResource `json:"incident_statuses"`
}

type incidentStatusWrapper struct {
	IncidentStatus IncidentStatusResource `json:"incident_status"`
}

func (c *Client) ListIncidentStatuses(ctx context.Context) ([]IncidentStatusResource, error) {
	result, err := doAndDecode[incidentStatusesWrapper](c, ctx, http.MethodGet, "/v1/incident_statuses", nil)
	if err != nil {
		return nil, err
	}
	return result.IncidentStatuses, nil
}

func (c *Client) GetIncidentStatus(ctx context.Context, id string) (*IncidentStatusResource, error) {
	return doAndDecodeField[incidentStatusWrapper, IncidentStatusResource](c, ctx, http.MethodGet, fmt.Sprintf("/v1/incident_statuses/%s", id), nil,
		func(w *incidentStatusWrapper) *IncidentStatusResource { return &w.IncidentStatus })
}
