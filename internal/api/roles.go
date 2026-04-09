package api

import (
	"context"
	"fmt"
	"net/http"
)

type IncidentRoleFull struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Required     bool   `json:"required"`
	Instructions string `json:"instructions"`
	Shortform    string `json:"shortform"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type incidentRolesWrapper struct {
	IncidentRoles []IncidentRoleFull `json:"incident_roles"`
}

type incidentRoleWrapper struct {
	IncidentRole IncidentRoleFull `json:"incident_role"`
}

func (c *Client) ListIncidentRoles(ctx context.Context) ([]IncidentRoleFull, error) {
	result, err := doAndDecode[incidentRolesWrapper](c, ctx, http.MethodGet, "/v2/incident_roles", nil)
	if err != nil {
		return nil, err
	}
	return result.IncidentRoles, nil
}

func (c *Client) GetIncidentRole(ctx context.Context, id string) (*IncidentRoleFull, error) {
	return doAndDecodeField[incidentRoleWrapper, IncidentRoleFull](c, ctx, http.MethodGet, fmt.Sprintf("/v2/incident_roles/%s", id), nil,
		func(w *incidentRoleWrapper) *IncidentRoleFull { return &w.IncidentRole })
}
