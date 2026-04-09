package api

import (
	"context"
	"fmt"
	"net/http"
)

type IncidentTimestampResource struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Rank int    `json:"rank"`
}

type incidentTimestampsWrapper struct {
	IncidentTimestamps []IncidentTimestampResource `json:"incident_timestamps"`
}

type incidentTimestampWrapper struct {
	IncidentTimestamp IncidentTimestampResource `json:"incident_timestamp"`
}

func (c *Client) ListIncidentTimestamps(ctx context.Context) ([]IncidentTimestampResource, error) {
	result, err := doAndDecode[incidentTimestampsWrapper](c, ctx, http.MethodGet, "/v2/incident_timestamps", nil)
	if err != nil {
		return nil, err
	}
	return result.IncidentTimestamps, nil
}

func (c *Client) GetIncidentTimestamp(ctx context.Context, id string) (*IncidentTimestampResource, error) {
	return doAndDecodeField[incidentTimestampWrapper, IncidentTimestampResource](c, ctx, http.MethodGet, fmt.Sprintf("/v2/incident_timestamps/%s", id), nil,
		func(w *incidentTimestampWrapper) *IncidentTimestampResource { return &w.IncidentTimestamp })
}
