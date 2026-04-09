package shared

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/shhac/agent-incident/internal/api"
	agenterrors "github.com/shhac/agent-incident/internal/errors"
)

// looksLikeID returns true if the input looks like an incident.io ULID
// (26 uppercase alphanumeric chars starting with 0-7).
var ulidPattern = regexp.MustCompile(`^[0-9A-HJ-NP-Za-hj-np-z]{26}$`)

func looksLikeID(s string) bool {
	return ulidPattern.MatchString(s)
}

// LooksLikeID is the exported version for use by other packages.
func LooksLikeID(s string) bool {
	return looksLikeID(s)
}

func ambiguousError(kind string, query string, names []string) error {
	return agenterrors.Newf(agenterrors.FixableByAgent,
		"ambiguous %s %q matched %d results", kind, query, len(names)).
		WithHint(fmt.Sprintf("Matches: %s. Use a more specific name or pass the ID directly.", strings.Join(names, ", ")))
}

func notFoundError(kind string, query string) error {
	return agenterrors.Newf(agenterrors.FixableByAgent,
		"no %s found matching %q", kind, query).
		WithHint(fmt.Sprintf("Use '%s list' to see available options.", kind))
}

// ResolveScheduleID resolves a schedule name-or-ID to an ID.
func ResolveScheduleID(ctx context.Context, client *api.Client, ref string) (string, error) {
	if looksLikeID(ref) {
		return ref, nil
	}
	schedules, err := client.ListSchedules(ctx)
	if err != nil {
		return "", err
	}
	return matchByName("schedule", ref, schedules, func(s api.Schedule) (string, string) { return s.ID, s.Name })
}

// ResolveEscalationPathID resolves an escalation path name-or-ID to an ID.
func ResolveEscalationPathID(ctx context.Context, client *api.Client, ref string) (string, error) {
	if looksLikeID(ref) {
		return ref, nil
	}
	paths, err := client.ListEscalationPaths(ctx)
	if err != nil {
		return "", err
	}
	return matchByName("escalation path", ref, paths, func(p api.EscalationPath) (string, string) { return p.ID, p.Name })
}

// ResolveStatusPageID resolves a status page name-or-ID to an ID.
func ResolveStatusPageID(ctx context.Context, client *api.Client, ref string) (string, error) {
	if looksLikeID(ref) {
		return ref, nil
	}
	pages, err := client.ListStatusPages(ctx)
	if err != nil {
		return "", err
	}
	return matchByName("status page", ref, pages, func(p api.StatusPage) (string, string) { return p.ID, p.Name })
}

// ResolveUserID resolves a user name, email, or ID to an ID.
func ResolveUserID(ctx context.Context, client *api.Client, ref string) (string, error) {
	if looksLikeID(ref) {
		return ref, nil
	}
	users, _, err := client.ListUsers(ctx, ref, 100, "")
	if err != nil {
		return "", err
	}

	lower := strings.ToLower(ref)
	var exact []api.User
	for _, u := range users {
		if strings.ToLower(u.Name) == lower || strings.ToLower(u.Email) == lower {
			exact = append(exact, u)
		}
	}

	if len(exact) == 1 {
		return exact[0].ID, nil
	}
	if len(exact) > 1 {
		names := make([]string, len(exact))
		for i, u := range exact {
			names[i] = fmt.Sprintf("%s <%s>", u.Name, u.Email)
		}
		return "", ambiguousError("user", ref, names)
	}

	if len(users) == 1 {
		return users[0].ID, nil
	}
	if len(users) > 1 {
		names := make([]string, len(users))
		for i, u := range users {
			names[i] = fmt.Sprintf("%s <%s>", u.Name, u.Email)
		}
		return "", ambiguousError("user", ref, names)
	}

	return "", notFoundError("user", ref)
}

func matchByName[T any](kind string, ref string, items []T, extract func(T) (string, string)) (string, error) {
	item, err := MatchByName(kind, ref, items, extract)
	if err != nil {
		return "", err
	}
	id, _ := extract(item)
	return id, nil
}

// MatchByName resolves a name to an item from a list, using exact-then-substring matching.
// Exported for use by domain packages that need the matched item, not just an ID.
func MatchByName[T any](kind string, ref string, items []T, extract func(T) (string, string)) (T, error) {
	lower := strings.ToLower(ref)

	var exactMatches []T
	var substringMatches []T
	for _, item := range items {
		_, name := extract(item)
		nameLower := strings.ToLower(name)
		if nameLower == lower {
			exactMatches = append(exactMatches, item)
		} else if strings.Contains(nameLower, lower) {
			substringMatches = append(substringMatches, item)
		}
	}

	if len(exactMatches) == 1 {
		return exactMatches[0], nil
	}
	if len(exactMatches) > 1 {
		var zero T
		return zero, ambiguousError(kind, ref, extractNames(exactMatches, extract))
	}

	if len(substringMatches) == 1 {
		return substringMatches[0], nil
	}
	if len(substringMatches) > 1 {
		var zero T
		return zero, ambiguousError(kind, ref, extractNames(substringMatches, extract))
	}

	var zero T
	return zero, notFoundError(kind, ref)
}

func extractNames[T any](items []T, extract func(T) (string, string)) []string {
	names := make([]string, len(items))
	for i, item := range items {
		_, name := extract(item)
		names[i] = name
	}
	return names
}

// ResolveSeverityID resolves a severity name-or-ID to an ID.
func ResolveSeverityID(ctx context.Context, client *api.Client, ref string) (string, error) {
	if looksLikeID(ref) {
		return ref, nil
	}
	severities, err := client.ListSeverities(ctx)
	if err != nil {
		return "", err
	}
	return matchByName("severity", ref, severities, func(s api.Severity) (string, string) { return s.ID, s.Name })
}

// ResolveIncidentStatusID resolves an incident status name-or-ID to an ID.
func ResolveIncidentStatusID(ctx context.Context, client *api.Client, ref string) (string, error) {
	if looksLikeID(ref) {
		return ref, nil
	}
	statuses, err := client.ListIncidentStatuses(ctx)
	if err != nil {
		return "", err
	}
	return matchByName("status", ref, statuses, func(s api.IncidentStatusResource) (string, string) { return s.ID, s.Name })
}

// ResolveCustomFieldID resolves a custom field name-or-ID to an ID.
func ResolveCustomFieldID(ctx context.Context, client *api.Client, ref string) (string, error) {
	if looksLikeID(ref) {
		return ref, nil
	}
	fields, err := client.ListCustomFields(ctx)
	if err != nil {
		return "", err
	}
	return matchByName("custom field", ref, fields, func(f api.CustomField) (string, string) { return f.ID, f.Name })
}

// ResolveIncidentTimestampID resolves a timestamp name-or-ID to an ID.
func ResolveIncidentTimestampID(ctx context.Context, client *api.Client, ref string) (string, error) {
	if looksLikeID(ref) {
		return ref, nil
	}
	timestamps, err := client.ListIncidentTimestamps(ctx)
	if err != nil {
		return "", err
	}
	return matchByName("timestamp", ref, timestamps, func(t api.IncidentTimestampResource) (string, string) { return t.ID, t.Name })
}

// ResolveCustomFieldOptionID resolves a custom field option value-or-ID to an ID
// for single/multi-select fields.
func ResolveCustomFieldOptionID(ref string, options []api.CustomFieldOption) (string, error) {
	if looksLikeID(ref) {
		return ref, nil
	}
	return matchByName("custom field option", ref, options, func(o api.CustomFieldOption) (string, string) { return o.ID, o.Value })
}

// ResolveCatalogEntryID resolves a catalog entry name-or-ID to an ID within a catalog type.
func ResolveCatalogEntryID(ctx context.Context, client *api.Client, catalogTypeID, ref string) (string, error) {
	if looksLikeID(ref) {
		return ref, nil
	}
	entries, _, err := client.ListCatalogEntries(ctx, catalogTypeID, ref, 25, "")
	if err != nil {
		return "", err
	}
	return matchByName("catalog entry", ref, entries, func(e api.CatalogEntry) (string, string) { return e.ID, e.Name })
}
