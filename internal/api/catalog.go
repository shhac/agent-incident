package api

import (
	"context"
	"net/http"
	"net/url"
)

type CatalogType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	TypeName    string `json:"type_name,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type CatalogEntry struct {
	ID            string                    `json:"id"`
	Name          string                    `json:"name"`
	CatalogTypeID string                   `json:"catalog_type_id"`
	AttributeValues map[string]CatalogAttrVal `json:"attribute_values,omitempty"`
	CreatedAt     string                    `json:"created_at"`
	UpdatedAt     string                    `json:"updated_at"`
}

type CatalogAttrVal struct {
	Value *CatalogAttrValue `json:"value,omitempty"`
}

type CatalogAttrValue struct {
	Literal string `json:"literal,omitempty"`
}

type catalogTypesWrapper struct {
	CatalogTypes []CatalogType `json:"catalog_types"`
}

type catalogTypeWrapper struct {
	CatalogType CatalogType `json:"catalog_type"`
}

type catalogEntriesWrapper struct {
	CatalogEntries []CatalogEntry  `json:"catalog_entries"`
	PaginationMeta *paginationMeta `json:"pagination_meta,omitempty"`
}

type catalogEntryWrapper struct {
	CatalogEntry CatalogEntry `json:"catalog_entry"`
}

func (c *Client) ListCatalogTypes(ctx context.Context) ([]CatalogType, error) {
	result, err := doAndDecode[catalogTypesWrapper](c, ctx, http.MethodGet, "/v2/catalog_types", nil)
	if err != nil {
		return nil, err
	}
	return result.CatalogTypes, nil
}

func (c *Client) GetCatalogType(ctx context.Context, id string) (*CatalogType, error) {
	return doAndDecodeField[catalogTypeWrapper, CatalogType](c, ctx, http.MethodGet, "/v2/catalog_types/"+id, nil,
		func(w *catalogTypeWrapper) *CatalogType { return &w.CatalogType })
}

func (c *Client) ListCatalogEntries(ctx context.Context, typeID, query string, pageSize int, after string) ([]CatalogEntry, string, error) {
	params := url.Values{}
	if typeID != "" {
		params.Set("catalog_type_id", typeID)
	}
	if query != "" {
		params.Set("name_contains", query)
	}
	addPaginationParams(params, pageSize, after)
	result, err := doAndDecode[catalogEntriesWrapper](c, ctx, http.MethodGet, buildPath("/v2/catalog_entries", params), nil)
	if err != nil {
		return nil, "", err
	}
	return result.CatalogEntries, extractCursor(result.PaginationMeta), nil
}

func (c *Client) GetCatalogEntry(ctx context.Context, id string) (*CatalogEntry, error) {
	return doAndDecodeField[catalogEntryWrapper, CatalogEntry](c, ctx, http.MethodGet, "/v2/catalog_entries/"+id, nil,
		func(w *catalogEntryWrapper) *CatalogEntry { return &w.CatalogEntry })
}
