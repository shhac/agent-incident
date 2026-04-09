package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

	agenterrors "github.com/shhac/agent-incident/internal/errors"
)

const defaultBaseURL = "https://api.incident.io"

type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
	debug   bool
}

func NewClient(apiKey string) *Client {
	return &Client{
		baseURL: defaultBaseURL,
		apiKey:  apiKey,
		http:    &http.Client{},
	}
}

func NewTestClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		http:    &http.Client{},
	}
}

// SetDebug enables HTTP request/response logging to stderr.
func (c *Client) SetDebug(enabled bool) {
	c.debug = enabled
}

func (c *Client) do(ctx context.Context, method, path string, body any) (json.RawMessage, error) {
	reqURL := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, agenterrors.Wrap(err, agenterrors.FixableByAgent)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return nil, agenterrors.Wrap(err, agenterrors.FixableByAgent)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, agenterrors.Wrap(err, agenterrors.FixableByRetry).WithHint("Network error — check connectivity")
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, agenterrors.Wrap(err, agenterrors.FixableByRetry)
	}

	if c.debug {
		c.logDebug(method, reqURL, resp.StatusCode, respBody)
	}

	if resp.StatusCode >= 400 {
		return nil, classifyHTTPError(resp.StatusCode, respBody)
	}

	return json.RawMessage(respBody), nil
}

func doAndDecode[T any](c *Client, ctx context.Context, method, path string, body any) (*T, error) {
	raw, err := c.do(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	var result T
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, agenterrors.Wrap(err, agenterrors.FixableByAgent)
	}
	return &result, nil
}

// doAndDecodeField decodes a JSON response and extracts a nested field.
func doAndDecodeField[W any, T any](c *Client, ctx context.Context, method, path string, body any, extract func(*W) *T) (*T, error) {
	wrapper, err := doAndDecode[W](c, ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	return extract(wrapper), nil
}

// paginationMeta is the shared pagination envelope used by incident.io list endpoints.
type paginationMeta struct {
	After        string `json:"after,omitempty"`
	TotalRecords int    `json:"total_record_count,omitempty"`
}

func extractCursor(meta *paginationMeta) string {
	if meta != nil {
		return meta.After
	}
	return ""
}

func addPaginationParams(params url.Values, pageSize int, after string) {
	if pageSize > 0 {
		params.Set("page_size", strconv.Itoa(pageSize))
	}
	if after != "" {
		params.Set("after", after)
	}
}

func buildPath(base string, params url.Values) string {
	if encoded := params.Encode(); encoded != "" {
		return base + "?" + encoded
	}
	return base
}

func extractErrorMessage(status int, body []byte) string {
	var parsed struct {
		Errors  []string `json:"errors"`
		Error   string   `json:"error"`
		Message string   `json:"message"`
	}
	if json.Unmarshal(body, &parsed) == nil {
		switch {
		case len(parsed.Errors) > 0 && parsed.Errors[0] != "":
			return parsed.Errors[0]
		case parsed.Error != "":
			return parsed.Error
		case parsed.Message != "":
			return parsed.Message
		}
	} else if len(body) > 0 && len(body) <= 200 {
		return fmt.Sprintf("HTTP %d: %s", status, string(body))
	}
	return fmt.Sprintf("HTTP %d", status)
}

func classifyHTTPError(status int, body []byte) *agenterrors.APIError {
	msg := extractErrorMessage(status, body)

	switch {
	case status == 401:
		return agenterrors.New("Authentication failed: "+msg, agenterrors.FixableByHuman).
			WithHint("Check your API key with 'agent-incident auth check'")
	case status == 403:
		return agenterrors.New("Permission denied: "+msg, agenterrors.FixableByHuman).
			WithHint("Your API key may not have sufficient permissions")
	case status == 404:
		return agenterrors.New("Not found: "+msg, agenterrors.FixableByAgent).
			WithHint("Check the ID — use 'list' to see available items")
	case status == 429:
		return agenterrors.New("Rate limited", agenterrors.FixableByRetry).
			WithHint("incident.io rate limit hit (1200 req/min) — wait and retry")
	case status >= 500:
		return agenterrors.New("incident.io API error: "+msg, agenterrors.FixableByRetry).
			WithHint("Server error — retry in a few seconds")
	default:
		return agenterrors.New(msg, agenterrors.FixableByAgent)
	}
}

func (c *Client) logDebug(method, url string, status int, body []byte) {
	entry := map[string]any{
		"@debug": "http",
		"method": method,
		"url":    url,
		"status": status,
	}
	var parsed any
	if json.Unmarshal(body, &parsed) == nil {
		entry["body"] = parsed
	} else {
		entry["body_raw"] = string(body)
	}
	enc := json.NewEncoder(os.Stderr)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	_ = enc.Encode(entry)
}
