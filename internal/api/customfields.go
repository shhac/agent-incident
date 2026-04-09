package api

import (
	"context"
	"net/http"
)

type CustomField struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	FieldType   string            `json:"field_type"`
	Required    string            `json:"required,omitempty"`
	Options     []CustomFieldOption `json:"options,omitempty"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
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
