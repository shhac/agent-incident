package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type IncidentStatusRef struct {
	Category string `json:"category"`
	Name     string `json:"name"`
}

type IncidentSeverity struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type IncidentRoleAssignment struct {
	Role     *IncidentRole `json:"role,omitempty"`
	Assignee *IncidentUser `json:"assignee,omitempty"`
}

type IncidentRole struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type IncidentUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

type CustomFieldEntry struct {
	CustomField *CustomFieldRef `json:"custom_field,omitempty"`
	Values      []CustomFieldValue `json:"values,omitempty"`
}

type CustomFieldRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CustomFieldValue struct {
	ValueLink     string `json:"value_link,omitempty"`
	ValueText     string `json:"value_text,omitempty"`
	ValueNumeric  string `json:"value_numeric,omitempty"`
	ValueCatalogEntry *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"value_catalog_entry,omitempty"`
}

type IncidentTimestamp struct {
	Name  string     `json:"name"`
	Value *time.Time `json:"value,omitempty"`
}

type ExternalResource struct {
	ResourceType string `json:"resource_type"`
	URL          string `json:"permalink"`
	Title        string `json:"title,omitempty"`
}

type IncidentType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Incident struct {
	ID                     string                   `json:"id"`
	Name                   string                   `json:"name"`
	Reference              string                   `json:"reference"`
	Status                 IncidentStatusRef           `json:"incident_status"`
	Severity               *IncidentSeverity         `json:"severity,omitempty"`
	Description            string                   `json:"description,omitempty"`
	Summary                string                   `json:"summary,omitempty"`
	CreatedAt              string                   `json:"created_at"`
	UpdatedAt              string                   `json:"updated_at"`
	IncidentRoleAssignments []IncidentRoleAssignment `json:"incident_role_assignments,omitempty"`
	CustomFieldEntries     []CustomFieldEntry        `json:"custom_field_entries,omitempty"`
	IncidentTimestamps     []IncidentTimestamp       `json:"incident_timestamps,omitempty"`
	ExternalResources      []ExternalResource        `json:"external_resources,omitempty"`
	IncidentType           *IncidentType             `json:"incident_type,omitempty"`
	Mode                   string                   `json:"mode,omitempty"`
	Visibility             string                   `json:"visibility,omitempty"`
	Permalink              string                   `json:"permalink,omitempty"`
}

type IncidentCompact struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	Severity     string `json:"severity,omitempty"`
	CreatedAt    string `json:"created_at"`
	IncidentLead string `json:"incident_lead,omitempty"`
}

func ToCompact(incidents []Incident) []IncidentCompact {
	result := make([]IncidentCompact, len(incidents))
	for i, inc := range incidents {
		result[i] = IncidentCompact{
			ID:        inc.ID,
			Name:      truncate(inc.Name, 200),
			Status:    inc.Status.Name,
			CreatedAt: inc.CreatedAt,
		}
		if inc.Severity != nil {
			result[i].Severity = inc.Severity.Name
		}
		for _, ra := range inc.IncidentRoleAssignments {
			if ra.Role != nil && ra.Role.Name == "Incident Lead" && ra.Assignee != nil {
				result[i].IncidentLead = ra.Assignee.Name
				break
			}
		}
	}
	return result
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

// ListIncidentsOpts configures the ListIncidents request.
type ListIncidentsOpts struct {
	StatusCategory []string
	Severity       []string
	PageSize       int
	After          string
}

type incidentsWrapper struct {
	Incidents        []Incident `json:"incidents"`
	PaginationMeta   *paginationMeta `json:"pagination_meta,omitempty"`
}

type incidentWrapper struct {
	Incident Incident `json:"incident"`
}


func (c *Client) ListIncidents(ctx context.Context, opts ListIncidentsOpts) ([]Incident, string, error) {
	params := url.Values{}
	for _, sc := range opts.StatusCategory {
		params.Add("status_category[]", sc)
	}
	for _, sev := range opts.Severity {
		params.Add("severity[]", sev)
	}
	addPaginationParams(params, opts.PageSize, opts.After)

	path := buildPath("/v2/incidents", params)
	resp, err := doAndDecode[incidentsWrapper](c, ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, "", err
	}

	return resp.Incidents, extractCursor(resp.PaginationMeta), nil
}

func (c *Client) GetIncident(ctx context.Context, id string) (*Incident, error) {
	path := fmt.Sprintf("/v2/incidents/%s", id)
	return doAndDecodeField[incidentWrapper, Incident](c, ctx, http.MethodGet, path, nil,
		func(w *incidentWrapper) *Incident { return &w.Incident })
}

// CreateIncidentParams are the parameters for creating an incident.
type CreateIncidentParams struct {
	Name       string            `json:"name"`
	Summary    string            `json:"summary,omitempty"`
	SeverityID string            `json:"severity_id,omitempty"`
	Mode       string            `json:"mode,omitempty"`
	Visibility string            `json:"visibility,omitempty"`
	IdempotencyKey string        `json:"idempotency_key"`
}

func (c *Client) CreateIncident(ctx context.Context, params CreateIncidentParams) (*Incident, error) {
	return doAndDecodeField[incidentWrapper, Incident](c, ctx, http.MethodPost, "/v2/incidents", params,
		func(w *incidentWrapper) *Incident { return &w.Incident })
}

// EditIncidentParams are the parameters for editing an incident.
type EditIncidentParams struct {
	Incident EditIncidentFields `json:"incident"`
}

type EditIncidentFields struct {
	Name       *string `json:"name,omitempty"`
	Summary    *string `json:"summary,omitempty"`
	SeverityID *string `json:"severity_id,omitempty"`
}

func (c *Client) EditIncident(ctx context.Context, id string, params EditIncidentParams) (*Incident, error) {
	path := fmt.Sprintf("/v2/incidents/%s/actions/edit", id)
	return doAndDecodeField[incidentWrapper, Incident](c, ctx, http.MethodPost, path, params,
		func(w *incidentWrapper) *Incident { return &w.Incident })
}

// IncidentUpdate represents an update posted to an incident.
type IncidentUpdate struct {
	ID         string `json:"id"`
	IncidentID string `json:"incident_id"`
	Message    string `json:"message,omitempty"`
	NewStatus  *IncidentStatusRef `json:"new_incident_status,omitempty"`
	CreatedAt  string `json:"created_at"`
	Updater    *IncidentUser `json:"updater,omitempty"`
}

type incidentUpdatesWrapper struct {
	IncidentUpdates []IncidentUpdate `json:"incident_updates"`
	PaginationMeta  *paginationMeta  `json:"pagination_meta,omitempty"`
}

func (c *Client) ListIncidentUpdates(ctx context.Context, incidentID string, pageSize int, after string) ([]IncidentUpdate, string, error) {
	params := url.Values{}
	params.Set("incident_id", incidentID)
	addPaginationParams(params, pageSize, after)

	path := buildPath("/v2/incident_updates", params)
	resp, err := doAndDecode[incidentUpdatesWrapper](c, ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, "", err
	}

	return resp.IncidentUpdates, extractCursor(resp.PaginationMeta), nil
}
