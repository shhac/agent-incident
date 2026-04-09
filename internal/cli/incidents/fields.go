package incidents

import (
	"context"
	"fmt"
	"strings"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/cli/shared"
)

func parseKeyValue(s string) (string, string, error) {
	idx := strings.IndexByte(s, '=')
	if idx < 0 {
		return "", "", fmt.Errorf("expected Name=Value format, got %q", s)
	}
	return s[:idx], s[idx+1:], nil
}

func resolveCustomFieldEntries(ctx context.Context, client *api.Client, flags []string) ([]api.EditCustomFieldEntry, error) {
	allFields, err := client.ListCustomFields(ctx)
	if err != nil {
		return nil, err
	}

	entries := make([]api.EditCustomFieldEntry, 0, len(flags))
	for _, flag := range flags {
		fieldName, value, err := parseKeyValue(flag)
		if err != nil {
			return nil, err
		}

		field, err := findCustomField(fieldName, allFields)
		if err != nil {
			return nil, err
		}

		entry := api.EditCustomFieldEntry{CustomFieldID: field.ID}
		if value == "" {
			entry.Values = []api.EditCustomFieldValue{}
		} else {
			v, err := buildFieldValue(ctx, client, field, value)
			if err != nil {
				return nil, err
			}
			entry.Values = []api.EditCustomFieldValue{v}
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func findCustomField(name string, fields []api.CustomField) (*api.CustomField, error) {
	field, err := shared.MatchByName("custom field", name, fields, func(f api.CustomField) (string, string) { return f.ID, f.Name })
	if err != nil {
		return nil, err
	}
	return &field, nil
}

func buildFieldValue(ctx context.Context, client *api.Client, field *api.CustomField, value string) (api.EditCustomFieldValue, error) {
	switch field.FieldType {
	case "single_select", "multi_select":
		if field.CatalogTypeID != "" {
			entryID, err := shared.ResolveCatalogEntryID(ctx, client, field.CatalogTypeID, value)
			if err != nil {
				return api.EditCustomFieldValue{}, fmt.Errorf("field %q: %w", field.Name, err)
			}
			return api.EditCustomFieldValue{ValueCatalogEntryID: entryID}, nil
		}
		options, _, err := client.ListCustomFieldOptions(ctx, field.ID, 250, "")
		if err != nil {
			return api.EditCustomFieldValue{}, fmt.Errorf("field %q: %w", field.Name, err)
		}
		optID, err := shared.ResolveCustomFieldOptionID(value, options)
		if err != nil {
			return api.EditCustomFieldValue{}, fmt.Errorf("field %q: %w", field.Name, err)
		}
		return api.EditCustomFieldValue{ValueOptionID: optID}, nil
	case "numeric":
		return api.EditCustomFieldValue{ValueNumeric: value}, nil
	case "link":
		return api.EditCustomFieldValue{ValueLink: value}, nil
	default:
		return api.EditCustomFieldValue{ValueText: value}, nil
	}
}

func resolveTimestampValues(ctx context.Context, client *api.Client, flags []string) ([]api.EditIncidentTimestampValue, error) {
	values := make([]api.EditIncidentTimestampValue, 0, len(flags))
	for _, flag := range flags {
		tsName, tsValue, err := parseKeyValue(flag)
		if err != nil {
			return nil, err
		}

		tsID, err := shared.ResolveIncidentTimestampID(ctx, client, tsName)
		if err != nil {
			return nil, err
		}

		entry := api.EditIncidentTimestampValue{IncidentTimestampID: tsID}
		if tsValue == "" {
			entry.Value = nil
		} else {
			t, err := shared.ParseTime(tsValue)
			if err != nil {
				return nil, fmt.Errorf("timestamp %q: %w", tsName, err)
			}
			iso := t.UTC().Format("2006-01-02T15:04:05Z")
			entry.Value = &iso
		}
		values = append(values, entry)
	}
	return values, nil
}
