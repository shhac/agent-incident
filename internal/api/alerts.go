package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Alert represents a full alert from the incident.io API.
type Alert struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	Description      string    `json:"description,omitempty"`
	Status           string    `json:"status"`
	DeduplicationKey string    `json:"deduplication_key,omitempty"`
	SourceURL        string    `json:"source_url,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// AlertCompact is a slim projection of Alert for list output.
type AlertCompact struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// ToCompactAlerts converts a slice of Alert to AlertCompact.
func ToCompactAlerts(alerts []Alert) []AlertCompact {
	out := make([]AlertCompact, len(alerts))
	for i, a := range alerts {
		out[i] = AlertCompact{
			ID:        a.ID,
			Title:     a.Title,
			Status:    a.Status,
			CreatedAt: a.CreatedAt,
		}
	}
	return out
}

// ListAlertsOpts holds optional parameters for ListAlerts.
type ListAlertsOpts struct {
	Status           []string
	DeduplicationKey string
	PageSize         int
	After            string
}

type alertsWrapper struct {
	Alerts         []Alert        `json:"alerts"`
	PaginationMeta paginationMeta `json:"pagination_meta"`
}

type alertWrapper struct {
	Alert Alert `json:"alert"`
}

// ListAlertsResult holds the alerts and pagination cursor from a list call.
type ListAlertsResult struct {
	Alerts []Alert
	After  string
}

// ListAlerts returns alerts matching the given filters.
func (c *Client) ListAlerts(ctx context.Context, opts ListAlertsOpts) (*ListAlertsResult, error) {
	params := url.Values{}
	for _, s := range opts.Status {
		params.Add("status[]", s)
	}
	if opts.DeduplicationKey != "" {
		params.Set("deduplication_key", opts.DeduplicationKey)
	}
	if opts.PageSize > 0 {
		params.Set("page_size", strconv.Itoa(opts.PageSize))
	}
	if opts.After != "" {
		params.Set("after", opts.After)
	}

	path := buildPath("/v2/alerts", params)
	w, err := doAndDecode[alertsWrapper](c, ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return &ListAlertsResult{
		Alerts: w.Alerts,
		After:  w.PaginationMeta.After,
	}, nil
}

// GetAlert retrieves a single alert by ID.
func (c *Client) GetAlert(ctx context.Context, id string) (*Alert, error) {
	return doAndDecodeField[alertWrapper, Alert](c, ctx, http.MethodGet, fmt.Sprintf("/v2/alerts/%s", id), nil,
		func(w *alertWrapper) *Alert { return &w.Alert })
}

// CreateAlertEventParams holds the body for creating an alert event.
type CreateAlertEventParams struct {
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	Status      string            `json:"status,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	DeduplicationKey string       `json:"deduplication_key,omitempty"`
}

type alertEventWrapper struct {
	Alert Alert `json:"alert"`
}

// CreateAlertEvent sends an alert event to the given alert source config.
func (c *Client) CreateAlertEvent(ctx context.Context, sourceConfigID string, params CreateAlertEventParams) (*Alert, error) {
	path := fmt.Sprintf("/v2/alert_events/http/%s", sourceConfigID)
	return doAndDecodeField[alertEventWrapper, Alert](c, ctx, http.MethodPost, path, params,
		func(w *alertEventWrapper) *Alert { return &w.Alert })
}

// IncidentAlert represents an alert attached to an incident.
type IncidentAlert struct {
	ID        string    `json:"id"`
	AlertID   string    `json:"alert_id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type incidentAlertsWrapper struct {
	IncidentAlerts []IncidentAlert `json:"incident_alerts"`
	PaginationMeta paginationMeta  `json:"pagination_meta"`
}

// ListIncidentAlertsResult holds the result of listing incident alerts.
type ListIncidentAlertsResult struct {
	IncidentAlerts []IncidentAlert
	After          string
}

// ListIncidentAlerts returns alerts that are attached to incidents.
func (c *Client) ListIncidentAlerts(ctx context.Context, pageSize int, after string) (*ListIncidentAlertsResult, error) {
	params := url.Values{}
	if pageSize > 0 {
		params.Set("page_size", strconv.Itoa(pageSize))
	}
	if after != "" {
		params.Set("after", after)
	}

	path := buildPath("/v2/incident_alerts", params)
	w, err := doAndDecode[incidentAlertsWrapper](c, ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return &ListIncidentAlertsResult{
		IncidentAlerts: w.IncidentAlerts,
		After:          w.PaginationMeta.After,
	}, nil
}
