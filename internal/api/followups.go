package api

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

type FollowUp struct {
	ID          string `json:"id"`
	IncidentID  string `json:"incident_id"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	Description string `json:"description,omitempty"`
	Assignee    *User  `json:"assignee,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	CompletedAt string `json:"completed_at,omitempty"`
}

type followUpsWrapper struct {
	FollowUps      []FollowUp      `json:"follow_ups"`
	PaginationMeta *paginationMeta `json:"pagination_meta,omitempty"`
}

type followUpWrapper struct {
	FollowUp FollowUp `json:"follow_up"`
}

func (c *Client) ListFollowUps(ctx context.Context, incidentID string, pageSize int, after string) ([]FollowUp, string, error) {
	params := url.Values{}
	if incidentID != "" {
		params.Set("incident_id", incidentID)
	}
	if pageSize > 0 {
		params.Set("page_size", strconv.Itoa(pageSize))
	}
	if after != "" {
		params.Set("after", after)
	}
	result, err := doAndDecode[followUpsWrapper](c, ctx, http.MethodGet, buildPath("/v2/follow_ups", params), nil)
	if err != nil {
		return nil, "", err
	}
	cursor := ""
	if result.PaginationMeta != nil {
		cursor = result.PaginationMeta.After
	}
	return result.FollowUps, cursor, nil
}

func (c *Client) GetFollowUp(ctx context.Context, id string) (*FollowUp, error) {
	return doAndDecodeField[followUpWrapper, FollowUp](c, ctx, http.MethodGet, "/v2/follow_ups/"+id, nil,
		func(w *followUpWrapper) *FollowUp { return &w.FollowUp })
}
