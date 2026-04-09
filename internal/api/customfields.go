package api

import (
	"context"
	"net/http"
	"net/url"
)

type CustomField struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	Description    string              `json:"description,omitempty"`
	FieldType      string              `json:"field_type"`
	Required       string              `json:"required,omitempty"`
	CatalogTypeID  string              `json:"catalog_type_id,omitempty"`
	Options        []CustomFieldOption `json:"options,omitempty"`
	CreatedAt      string              `json:"created_at"`
	UpdatedAt      string              `json:"updated_at"`
}

type CustomFieldOption struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

type customFieldsWrapper struct {
	CustomFields []CustomField `json:"custom_fields"`
}

type customFieldWrapper struct {
	CustomField CustomField `json:"custom_field"`
}

func (c *Client) ListCustomFields(ctx context.Context) ([]CustomField, error) {
	result, err := doAndDecode[customFieldsWrapper](c, ctx, http.MethodGet, "/v2/custom_fields", nil)
	if err != nil {
		return nil, err
	}
	return result.CustomFields, nil
}

func (c *Client) GetCustomField(ctx context.Context, id string) (*CustomField, error) {
	return doAndDecodeField[customFieldWrapper, CustomField](c, ctx, http.MethodGet, "/v2/custom_fields/"+id, nil,
		func(w *customFieldWrapper) *CustomField { return &w.CustomField })
}

type customFieldOptionsWrapper struct {
	CustomFieldOptions []CustomFieldOption `json:"custom_field_options"`
	PaginationMeta     *paginationMeta    `json:"pagination_meta,omitempty"`
}

func (c *Client) ListCustomFieldOptions(ctx context.Context, customFieldID string, pageSize int, after string) ([]CustomFieldOption, string, error) {
	params := url.Values{}
	params.Set("custom_field_id", customFieldID)
	addPaginationParams(params, pageSize, after)

	path := buildPath("/v1/custom_field_options", params)
	resp, err := doAndDecode[customFieldOptionsWrapper](c, ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, "", err
	}
	return resp.CustomFieldOptions, extractCursor(resp.PaginationMeta), nil
}
