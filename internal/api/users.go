package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type User struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	SlackUserID string `json:"slack_user_id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type UserCompact struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func ToCompactUsers(users []User) []UserCompact {
	result := make([]UserCompact, len(users))
	for i, u := range users {
		result[i] = UserCompact{
			ID:    u.ID,
			Name:  u.Name,
			Email: u.Email,
			Role:  u.Role,
		}
	}
	return result
}

type usersWrapper struct {
	Users          []User         `json:"users"`
	PaginationMeta paginationMeta `json:"pagination_meta"`
}

type userWrapper struct {
	User User `json:"user"`
}

func (c *Client) ListUsers(ctx context.Context, query string, pageSize int, after string) ([]User, paginationMeta, error) {
	params := url.Values{}
	if query != "" {
		params.Set("query", query)
	}
	addPaginationParams(params, pageSize, after)
	path := buildPath("/v2/users", params)
	result, err := doAndDecode[usersWrapper](c, ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, paginationMeta{}, err
	}
	return result.Users, result.PaginationMeta, nil
}

func (c *Client) GetUser(ctx context.Context, id string) (*User, error) {
	return doAndDecodeField[userWrapper, User](c, ctx, http.MethodGet, fmt.Sprintf("/v2/users/%s", id), nil,
		func(w *userWrapper) *User { return &w.User })
}
