package api

import (
	"context"
	"fmt"
	"net/http"
)

type Escalation struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type EscalationPath struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path []any  `json:"path"`
}

type CreateEscalationParams struct {
	IncidentID       string `json:"incident_id"`
	EscalationPathID string `json:"escalation_path_id"`
}

type escalationsWrapper struct {
	Escalations []Escalation `json:"escalations"`
}

type escalationWrapper struct {
	Escalation Escalation `json:"escalation"`
}

type escalationPathsWrapper struct {
	EscalationPaths []EscalationPath `json:"escalation_paths"`
}

type escalationPathWrapper struct {
	EscalationPath EscalationPath `json:"escalation_path"`
}

func (c *Client) ListEscalations(ctx context.Context) ([]Escalation, error) {
	resp, err := doAndDecode[escalationsWrapper](c, ctx, http.MethodGet, "/v2/escalations", nil)
	if err != nil {
		return nil, err
	}
	return resp.Escalations, nil
}

func (c *Client) GetEscalation(ctx context.Context, id string) (*Escalation, error) {
	path := fmt.Sprintf("/v2/escalations/%s", id)
	return doAndDecodeField[escalationWrapper, Escalation](c, ctx, http.MethodGet, path, nil,
		func(w *escalationWrapper) *Escalation { return &w.Escalation })
}

func (c *Client) CreateEscalation(ctx context.Context, params CreateEscalationParams) (*Escalation, error) {
	return doAndDecodeField[escalationWrapper, Escalation](c, ctx, http.MethodPost, "/v2/escalations", params,
		func(w *escalationWrapper) *Escalation { return &w.Escalation })
}

func (c *Client) ListEscalationPaths(ctx context.Context) ([]EscalationPath, error) {
	resp, err := doAndDecode[escalationPathsWrapper](c, ctx, http.MethodGet, "/v2/escalation_paths", nil)
	if err != nil {
		return nil, err
	}
	return resp.EscalationPaths, nil
}

func (c *Client) GetEscalationPath(ctx context.Context, id string) (*EscalationPath, error) {
	path := fmt.Sprintf("/v2/escalation_paths/%s", id)
	return doAndDecodeField[escalationPathWrapper, EscalationPath](c, ctx, http.MethodGet, path, nil,
		func(w *escalationPathWrapper) *EscalationPath { return &w.EscalationPath })
}
