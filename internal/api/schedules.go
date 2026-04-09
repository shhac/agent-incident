package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Schedule struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Timezone  string `json:"timezone"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ScheduleEntryUser struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	SlackUserID string `json:"slack_user_id,omitempty"`
}

type ScheduleEntry struct {
	RotationID  string            `json:"rotation_id,omitempty"`
	Fingerprint string            `json:"fingerprint,omitempty"`
	User        ScheduleEntryUser `json:"user"`
	StartAt     string            `json:"start_at"`
	EndAt       string            `json:"end_at"`
}

type ScheduleOverrideParams struct {
	ScheduleID string `json:"schedule_id"`
	UserID     string `json:"user_id"`
	StartAt    string `json:"start_at"`
	EndAt      string `json:"end_at"`
}

type ScheduleOverride struct {
	ID         string `json:"id"`
	ScheduleID string `json:"schedule_id"`
	UserID     string `json:"user_id"`
	StartAt    string `json:"start_at"`
	EndAt      string `json:"end_at"`
}

type schedulesWrapper struct {
	Schedules []Schedule `json:"schedules"`
}

type scheduleWrapper struct {
	Schedule Schedule `json:"schedule"`
}

type scheduleEntriesGroup struct {
	Final     []ScheduleEntry `json:"final"`
	Overrides []ScheduleEntry `json:"overrides"`
	Scheduled []ScheduleEntry `json:"scheduled"`
}

type scheduleEntriesWrapper struct {
	ScheduleEntries scheduleEntriesGroup `json:"schedule_entries"`
}

type scheduleOverrideWrapper struct {
	ScheduleOverride ScheduleOverride `json:"schedule_override"`
}

func (c *Client) ListSchedules(ctx context.Context) ([]Schedule, error) {
	resp, err := doAndDecode[schedulesWrapper](c, ctx, http.MethodGet, "/v2/schedules", nil)
	if err != nil {
		return nil, err
	}
	return resp.Schedules, nil
}

func (c *Client) GetSchedule(ctx context.Context, id string) (*Schedule, error) {
	path := fmt.Sprintf("/v2/schedules/%s", id)
	return doAndDecodeField[scheduleWrapper, Schedule](c, ctx, http.MethodGet, path, nil,
		func(w *scheduleWrapper) *Schedule { return &w.Schedule })
}

func (c *Client) ListScheduleEntries(ctx context.Context, scheduleID string, from, to time.Time) ([]ScheduleEntry, error) {
	params := url.Values{}
	params.Set("schedule_id", scheduleID)
	params.Set("entry_window_start", from.Format(time.RFC3339))
	params.Set("entry_window_end", to.Format(time.RFC3339))

	path := buildPath("/v2/schedule_entries", params)
	resp, err := doAndDecode[scheduleEntriesWrapper](c, ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return resp.ScheduleEntries.Final, nil
}

func (c *Client) CreateScheduleOverride(ctx context.Context, params ScheduleOverrideParams) (*ScheduleOverride, error) {
	return doAndDecodeField[scheduleOverrideWrapper, ScheduleOverride](c, ctx, http.MethodPost, "/v2/schedule_overrides", params,
		func(w *scheduleOverrideWrapper) *ScheduleOverride { return &w.ScheduleOverride })
}
