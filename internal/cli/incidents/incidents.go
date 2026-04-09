package incidents

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/cli/shared"
)

// normalizeIncidentRef strips a "INC-" prefix when followed by digits,
// so users can pass "INC-2000" or "2000" as an incident reference.
// UUIDs and other ID formats are passed through unchanged.
func normalizeIncidentRef(ref string) string {
	upper := strings.ToUpper(ref)
	if strings.HasPrefix(upper, "INC-") {
		suffix := ref[4:]
		if _, err := strconv.Atoi(suffix); err == nil {
			return suffix
		}
	}
	return ref
}

// resolveIncidentID converts a reference (INC-2000, 2000) to a UUID by fetching the incident.
// If the input is already a UUID-length string, it's returned as-is.
func resolveIncidentID(ctx context.Context, client *api.Client, ref string) (string, error) {
	normalized := normalizeIncidentRef(ref)
	incident, err := client.GetIncident(ctx, normalized)
	if err != nil {
		return "", err
	}
	return incident.ID, nil
}

// Register adds the incidents command group to the root command.
func Register(root *cobra.Command, globals shared.GlobalsFunc) {
	incidents := &cobra.Command{
		Use:   "incident",
		Short: "Manage incidents",
	}

	registerList(incidents, globals)
	registerGet(incidents, globals)
	registerCreate(incidents, globals)
	registerEdit(incidents, globals)
	registerUpdates(incidents, globals)
	shared.RegisterLLMHelp(incidents, "LLM reference for incidents commands", incidentsLLMHelp)

	root.AddCommand(incidents)
}

func registerList(parent *cobra.Command, globals shared.GlobalsFunc) {
	var (
		status   []string
		severity []string
		from  string
		to    string
		limit    int
		after    string
		full     bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List incidents",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()

			var createdAfter, createdBefore string
			if from != "" {
				t, err := shared.ParseTime(from)
				if err != nil {
					return err
				}
				createdAfter = t.Format("2006-01-02")
			}
			if to != "" {
				t, err := shared.ParseTime(to)
				if err != nil {
					return err
				}
				createdBefore = t.Format("2006-01-02")
			}

			return shared.WithClient(g, func(ctx context.Context, client *api.Client) error {
				opts := api.ListIncidentsOpts{
					StatusCategory: status,
					Severity:       severity,
					CreatedAfter:   createdAfter,
					CreatedBefore:  createdBefore,
					PageSize:       limit,
					After:          after,
				}

				incidents, cursor, err := client.ListIncidents(ctx, opts)
				if err != nil {
					return err
				}

				pagination := shared.CursorPagination(cursor)

				if full {
					shared.WritePaginatedList(shared.ToAnySlice(incidents), pagination, g.Format)
				} else {
					compact := api.ToCompact(incidents)
					shared.WritePaginatedList(shared.ToAnySlice(compact), pagination, g.Format)
				}
				return nil
			})
		},
	}

	cmd.Flags().StringSliceVar(&status, "status", nil, "Filter by status category (active, closed, etc.)")
	cmd.Flags().StringSliceVar(&severity, "severity", nil, "Filter by severity name")
	cmd.Flags().StringVar(&from, "from", "", "Only show incidents created after this time")
	cmd.Flags().StringVar(&to, "to", "", "Only show incidents created before this time")
	cmd.Flags().IntVar(&limit, "limit", 25, "Number of results per page")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor")
	cmd.Flags().BoolVar(&full, "full", false, "Return full incident objects instead of compact")
	parent.AddCommand(cmd)
}

func registerGet(parent *cobra.Command, globals shared.GlobalsFunc) {
	cmd := &cobra.Command{
		Use:   "get <id-or-reference>",
		Short: "Get a single incident by ID or reference (e.g. INC-2000 or 2000)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g, func(ctx context.Context, client *api.Client) error {
				id := normalizeIncidentRef(args[0])
				incident, err := client.GetIncident(ctx, id)
				if err != nil {
					return err
				}
				shared.WriteItem(incident, g.Format)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
}

func registerCreate(parent *cobra.Command, globals shared.GlobalsFunc) {
	var (
		name     string
		severity string
		summary  string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new incident",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !shared.RequireFlag("name", name, "Provide --name for the incident") {
				return nil
			}
			g := globals()
			return shared.WithClient(g, func(ctx context.Context, client *api.Client) error {
				params := api.CreateIncidentParams{
					Name:           name,
					Summary:        summary,
					SeverityID:     severity,
					IdempotencyKey: shared.NewIdempotencyKey(),
				}
				incident, err := client.CreateIncident(ctx, params)
				if err != nil {
					return err
				}
				shared.WriteItem(incident, g.Format)
				return nil
			})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Incident name (required)")
	cmd.Flags().StringVar(&severity, "severity", "", "Severity ID")
	cmd.Flags().StringVar(&summary, "summary", "", "Incident summary")
	parent.AddCommand(cmd)
}

func registerEdit(parent *cobra.Command, globals shared.GlobalsFunc) {
	var (
		name       string
		severity   string
		status     string
		summary    string
		fieldFlags []string
		tsFlags    []string
	)

	cmd := &cobra.Command{
		Use:   "edit <id-or-reference>",
		Short: "Edit an existing incident (accepts INC-2000, 2000, or UUID)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g, func(ctx context.Context, client *api.Client) error {
				fields := api.EditIncidentFields{}
				if cmd.Flags().Changed("name") {
					fields.Name = &name
				}
				if cmd.Flags().Changed("summary") {
					fields.Summary = &summary
				}

				if cmd.Flags().Changed("severity") {
					sevID, err := shared.ResolveSeverityID(ctx, client, severity)
					if err != nil {
						return err
					}
					fields.SeverityID = &sevID
				}

				if cmd.Flags().Changed("status") {
					statusID, err := shared.ResolveIncidentStatusID(ctx, client, status)
					if err != nil {
						return err
					}
					fields.IncidentStatusID = &statusID
				}

				if len(fieldFlags) > 0 {
					entries, err := resolveCustomFieldEntries(ctx, client, fieldFlags)
					if err != nil {
						return err
					}
					fields.CustomFieldEntries = entries
				}

				if len(tsFlags) > 0 {
					tsValues, err := resolveTimestampValues(ctx, client, tsFlags)
					if err != nil {
						return err
					}
					fields.IncidentTimestampValues = tsValues
				}

				params := api.EditIncidentParams{Incident: fields}
				id, err := resolveIncidentID(ctx, client, args[0])
				if err != nil {
					return err
				}
				incident, err := client.EditIncident(ctx, id, params)
				if err != nil {
					return err
				}
				shared.WriteItem(incident, g.Format)
				return nil
			})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "New incident name")
	cmd.Flags().StringVar(&severity, "severity", "", "Severity (name or ID)")
	cmd.Flags().StringVar(&status, "status", "", "Incident status (name or ID)")
	cmd.Flags().StringVar(&summary, "summary", "", "Updated summary")
	cmd.Flags().StringArrayVar(&fieldFlags, "field", nil, `Set custom field value (repeatable): "Field Name=value"`)
	cmd.Flags().StringArrayVar(&tsFlags, "timestamp", nil, `Set timestamp value (repeatable): "Reported at=2026-04-09T15:00:00Z"`)
	parent.AddCommand(cmd)
}

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

func registerUpdates(parent *cobra.Command, globals shared.GlobalsFunc) {
	var (
		limit int
		after string
	)

	cmd := &cobra.Command{
		Use:   "updates <id>",
		Short: "List updates for an incident",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g, func(ctx context.Context, client *api.Client) error {
				incidentID, err := resolveIncidentID(ctx, client, args[0])
				if err != nil {
					return err
				}
				updates, cursor, err := client.ListIncidentUpdates(ctx, incidentID, limit, after)
				if err != nil {
					return err
				}
				pagination := shared.CursorPagination(cursor)
				shared.WritePaginatedList(shared.ToAnySlice(updates), pagination, g.Format)
				return nil
			})
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 25, "Number of results per page")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor")
	parent.AddCommand(cmd)
}

