# agent-incident: Implementation Tasks

Reference: [Design Doc](design-docs/commands-and-domains.md) | [OpenAPI Spec](incident-openapi-v3.json)

API base: `https://api.incident.io` | Auth: `Authorization: Bearer <api-key>` | Identity check: `GET /v1/identity`

---

## Phase 0: Scaffolding

### 1. Initialize Go module and project skeleton
- [ ] **Status:** Not started
- **Description:** Create `go.mod` (`github.com/shhac/agent-incident`, Go 1.23), add `github.com/spf13/cobra` and `gopkg.in/yaml.v3` dependencies. Create directory structure: `cmd/agent-incident/`, `internal/cli/`, `internal/api/`, `internal/config/`, `internal/credential/`, `internal/errors/`, `internal/output/`, `internal/cli/shared/`.
- **Files:** `go.mod`, all directories above (populated in later tasks)
- **Dependencies:** None

### 2. Create `cmd/agent-incident/main.go`
- [ ] **Status:** Not started
- **Description:** Entry point matching sibling pattern. Accepts `version` via ldflags, calls `cli.Execute(version)`, exits with code 1 on error.
- **Files:** `cmd/agent-incident/main.go`
- **Dependencies:** 1

### 3. Create `internal/cli/root.go` with root command and global flags
- [ ] **Status:** Not started
- **Description:** Root cobra command (`agent-incident`). Persistent flags: `--organization` (org alias), `--api-key` (direct key override), `--format` (json/yaml/jsonl), `--timeout` (ms). Wire `Execute(version)` function. Include `SilenceUsage: true`, `SilenceErrors: true`. Add `version` subcommand inline. Include `allGlobals()` helper returning `*shared.GlobalFlags` (matches agent-dd pattern). Domain commands receive `globals func() *shared.GlobalFlags`; auth receives only `root *cobra.Command`.
- **Files:** `internal/cli/root.go`
- **Dependencies:** 2

### 4. Create Makefile
- [ ] **Status:** Not started
- **Description:** Targets: `build`, `test`, `test-short`, `lint`, `fmt`, `clean`, `dev`, `vet`. Binary name `agent-incident`. Ldflags inject `main.version`. Match agent-dd Makefile structure.
- **Files:** `Makefile`
- **Dependencies:** 2

### 5. Create `.goreleaser.yml`
- [ ] **Status:** Not started
- **Description:** Cross-compile for darwin/linux (amd64+arm64) + windows/amd64 (skip windows/arm64). Binary name `agent-incident`. CGO_ENABLED=0. tar.gz archives (zip for windows), sha256 checksums. GitHub release to `shhac/agent-incident`. Include changelog section with sort and exclude filters (chore/docs/test) matching agent-dd.
- **Files:** `.goreleaser.yml`
- **Dependencies:** 1

### 6. Create `.gitignore`
- [ ] **Status:** Not started
- **Description:** Ignore: `/agent-incident`, `vendor/`, `dist/`, `release/`, `.DS_Store`, `.env`, `.env.*`, `.cache`, `.tmp`, `.ai-cache`, `.claude/*.local.json`, `CLAUDE.local.md`, `design-docs/`.
- **Files:** `.gitignore` (update existing)
- **Dependencies:** None

### 7. Create `CLAUDE.md` and `AGENTS.md`
- [ ] **Status:** Not started
- **Description:** `CLAUDE.md`: Project-specific instructions for AI agents. Cover: project purpose (incident.io triage CLI for AI agents), dev workflow (`make build/test/dev`), testing patterns (mock server via `shared.ClientFactory`), architecture overview (cobra commands -> shared.WithClient -> api.Client -> output), environment variables (`INCIDENT_API_KEY`, `INCIDENT_API_URL`), sibling project reference (agent-dd for patterns). Include note about incident.io API v2 being the primary version. `AGENTS.md`: Subagent role definitions matching agent-dd pattern (reviewer, implementer, etc.).
- **Files:** `CLAUDE.md`, `AGENTS.md`
- **Dependencies:** 1

---

## Phase 0.5: Shared Infrastructure

### 8. Implement `internal/errors/errors.go`
- [ ] **Status:** Not started
- **Description:** Structured error types matching agent-dd pattern. `APIError` with `Message`, `Hint`, `FixableBy` (agent/human/retry), `Cause`. Constructors: `New`, `Newf`, `Wrap`. Methods: `WithHint`, `WithCause`. Helper: `As` (wraps `errors.As`).
- **Files:** `internal/errors/errors.go`
- **Dependencies:** 1

### 9. Implement `internal/output/output.go`
- [ ] **Status:** Not started
- **Description:** Output formatting: `FormatJSON`, `FormatYAML`, `FormatNDJSON`. Functions: `ParseFormat`, `ResolveFormat`, `Print` (dispatches by format), `PrintJSON`, `PrintJSONList` (wraps items in `{"data": [...]}` envelope per design doc), `WriteError` (structured error JSON to stderr). `NDJSONWriter` with `WriteItem` and `WritePagination`. `Pagination` struct with `HasMore`, `TotalItems`, `NextCursor`. Null pruning via `pruneNulls` recursion. All JSON output uses 2-space indent, no HTML escaping.
- **Files:** `internal/output/output.go`
- **Dependencies:** 8

### 10. Implement `internal/config/config.go`
- [ ] **Status:** Not started
- **Description:** Config file at `~/.config/agent-incident/config.json`. `Config` struct with `DefaultOrg`, `Organizations` map, `Settings`. Thread-safe caching with mutex. Functions: `ConfigDir` (respects `XDG_CONFIG_HOME`), `Read`, `Write`, `ClearCache`, `SetConfigDir` (for tests), `StoreOrganization`, `RemoveOrganization`, `SetDefault`. Organization struct has no `Site` field (unlike agent-dd, incident.io has a single API host).
- **Files:** `internal/config/config.go`
- **Dependencies:** 1

### 11. Implement `internal/credential/credential.go`
- [ ] **Status:** Not started
- **Description:** Credential storage at `~/.config/agent-incident/credentials.json`. `Credential` struct with `APIKey` and `KeychainManaged`. Simpler than agent-dd (single key, no app key). Functions: `Store` (try keychain first, fall back to file), `Get`, `Remove`, `List`. `NotFoundError` type. File permissions `0o600` for credentials.
- **Files:** `internal/credential/credential.go`
- **Dependencies:** 10

### 12. Implement `internal/credential/keychain.go`
- [ ] **Status:** Not started
- **Description:** macOS keychain integration via `security` CLI. Service name `app.paulie.agent-incident`. Functions: `keychainStore`, `keychainGet`, `keychainDelete`. Only stores single API key (not api_key+app_key pair like agent-dd). Non-darwin platforms return error, causing graceful fallback to file storage.
- **Files:** `internal/credential/keychain.go`
- **Dependencies:** 1 (built alongside task 11, same package)

### 13. Implement `internal/api/client.go`
- [ ] **Status:** Not started
- **Description:** HTTP client for incident.io API. `Client` struct with `baseURL`, `apiKey`, `http`. Default base URL `https://api.incident.io`. Auth via `Authorization: Bearer <apiKey>` header. Core `do` method handling request building, error classification, response reading. Generic helpers: `doAndDecode[T]`, `doAndDecodeField[W,T]`. `buildPath` for query params. `classifyHTTPError` mapping 401/403/404/429/5xx to appropriate `FixableBy` categories with incident.io-specific hints. `NewClient(apiKey)`, `NewTestClient(baseURL, apiKey)`.
- **Files:** `internal/api/client.go`
- **Dependencies:** 8

### 14. Implement `internal/cli/shared/shared.go`
- [ ] **Status:** Not started
- **Description:** Shared CLI utilities. `GlobalFlags` struct (`Org`, `APIKey`, `Format`, `Timeout`). `GlobalsFunc` type alias (`func() *GlobalFlags`) for domain Register signatures. `MakeContext` (timeout support). `ResolveOrg` (resolution order: org flag -> `INCIDENT_API_KEY` env shortcircuit -> config default). `NewClientFromOrg` (resolution: `--api-key` flag -> `INCIDENT_API_KEY` env -> `INCIDENT_API_URL` test override -> credential store lookup). `WithClient` pattern (resolve client, run callback, write errors to stderr). Helpers: `WritePaginatedList` (default format NDJSON), `WriteItem` (default format JSON), `CursorPagination`, `RequireFlag`, `ToAnySlice`. `ClientFactory` var for test injection.
- **Files:** `internal/cli/shared/shared.go`
- **Dependencies:** 9, 10, 11, 13

### 15. Implement `internal/cli/shared/llmhelp.go`
- [ ] **Status:** Not started
- **Description:** `RegisterLLMHelp` function to add `llm-help` subcommand to any domain command. Prints domain-specific reference text. Matches agent-dd pattern.
- **Files:** `internal/cli/shared/llmhelp.go`
- **Dependencies:** 14

### 16. Implement `internal/cli/usage.go` with top-level `llm-help` command
- [ ] **Status:** Not started
- **Description:** Register a root-level `llm-help` command that prints a comprehensive reference card covering all domains, global flags, auth setup, and output formats. Update as new domains are added.
- **Files:** `internal/cli/usage.go`
- **Dependencies:** 3

### 17. Implement `internal/cli/shared/testhelper.go`
- [ ] **Status:** Not started
- **Description:** `SetupMockServer` function that creates an `httptest.Server` and injects it via `shared.ClientFactory`. Used by all domain command tests. Returns server and cleanup function.
- **Files:** `internal/cli/shared/testhelper.go`
- **Dependencies:** 14

### 18. Implement `internal/cli/shared/timeparse.go` for relative time parsing
- [ ] **Status:** Not started
- **Description:** Parse relative time strings — both backward (`now-15m`, `now-1h`, `now-1d`, `now-7d`) and forward (`now+1h`, `now+30m`, `now+1d`) — RFC3339 absolute times, and Unix epoch seconds. Forward-looking times are needed for `--to` flags (e.g., `schedules entries --to now+1h`). Used by `--since`, `--from`, `--to` flags across incidents, alerts, schedules, and other domains. Match agent-dd `timeparse.go` pattern.
- **Files:** `internal/cli/shared/timeparse.go`, `internal/cli/shared/timeparse_test.go`
- **Dependencies:** 14

### 19. Add unit tests for shared infrastructure
- [ ] **Status:** Not started
- **Description:** Tests for: `internal/errors/errors_test.go` (error construction, wrapping, hints, As), `internal/output/output_test.go` (JSON/YAML/NDJSON formatting, null pruning, error output), `internal/config/config_test.go` (read/write/default/remove with temp dirs), `internal/credential/credential_test.go` (store/get/remove/list with temp dirs, file-only mode), `internal/api/client_test.go` (HTTP error classification, buildPath, doAndDecode with mock server), `internal/cli/shared/timeparse_test.go` (relative, absolute, epoch parsing).
- **Files:** `internal/errors/errors_test.go`, `internal/output/output_test.go`, `internal/config/config_test.go`, `internal/credential/credential_test.go`, `internal/api/client_test.go`
- **Dependencies:** 8, 9, 10, 11, 13, 17, 18

---

## Phase 1: Core Triage Loop

### 20. Implement `auth` commands — `internal/cli/auth/auth.go`
- [ ] **Status:** Not started
- **Description:** Auth domain with subcommands: `add <alias>` (prompt-free: `--api-key` flag), `check [alias]` (calls `GET /v1/identity` to verify key), `default <alias>`, `list` (show all orgs, mark default), `remove <alias>`. Register on root command. Wire to config + credential packages.
- **Files:** `internal/cli/auth/auth.go`
- **Dependencies:** 3, 10, 11, 14

### 21. Implement `GET /v1/identity` in API client
- [ ] **Status:** Not started
- **Description:** `Client.GetIdentity(ctx)` method that calls `GET /v1/identity` and returns the identity response. Used by `auth check` to verify credentials.
- **Files:** `internal/api/identity.go`
- **Dependencies:** 13

### 22. Add tests for `auth` commands
- [ ] **Status:** Not started
- **Description:** Test all auth subcommands: add stores credential and config, check calls identity endpoint, default switches org, list shows all orgs, remove deletes credential. Use temp dirs to avoid touching real config.
- **Files:** `internal/cli/auth/auth_test.go`
- **Dependencies:** 20, 21, 17

### 23. Implement `incidents` API methods — `internal/api/incidents.go`
- [ ] **Status:** Not started
- **Description:** Types: `Incident` (full), `IncidentCompact` (id, name, status, severity, created_at, incident_lead). API methods: `ListIncidents(ctx, status, severity, since, limit, cursor)` using `GET /v2/incidents`, `GetIncident(ctx, id)` using `GET /v2/incidents/{id}`, `CreateIncident(ctx, params)` using `POST /v2/incidents`, `EditIncident(ctx, id, params)` using `POST /v2/incidents/{id}/actions/edit`. Also `ListIncidentUpdates(ctx, incidentID)` using `GET /v2/incident_updates` with incident_id filter. Compact projection: truncate description/summary to 200 chars, flatten role assignments to `{role: user}`, omit custom_field_entries/timestamps/external_resources.
- **Files:** `internal/api/incidents.go`
- **Dependencies:** 13

### 24. Implement `incidents` CLI commands — `internal/cli/incidents/incidents.go`
- [ ] **Status:** Not started
- **Description:** Register `incidents` domain with subcommands: `list` (flags: `--status`, `--severity`, `--since`, `--limit`, `--after` cursor, `--full`), `get <id>`, `create` (flags: `--name`, `--severity`, `--summary`), `edit <id>` (flags: `--status`, `--severity`, `--summary`), `updates <id>`. Add `llm-help` subcommand. Compact projection on `list` by default, full on `--full` or `get`. Uses timeparse for `--since`.
- **Files:** `internal/cli/incidents/incidents.go`, `internal/cli/incidents/usage.go`
- **Dependencies:** 14, 15, 18, 23

### 25. Add tests for `incidents` commands
- [ ] **Status:** Not started
- **Description:** Test list (with filters, compact vs full), get, create, edit, updates. Verify correct API paths, methods, query params, request bodies. Use mock server.
- **Files:** `internal/cli/incidents/incidents_test.go`
- **Dependencies:** 24, 17

### 26. Implement `alerts` API methods — `internal/api/alerts.go`
- [ ] **Status:** Not started
- **Description:** Types: `Alert`, `AlertCompact`. API methods: `ListAlerts(ctx, status, source, since, limit, cursor)` using `GET /v2/alerts`, `GetAlert(ctx, id)` using `GET /v2/alerts/{id}`, `CreateAlertEvent(ctx, sourceConfigID, params)` using `POST /v2/alert_events/http/{alert_source_config_id}`, `ListIncidentAlerts(ctx)` using `GET /v2/incident_alerts`.
- **Files:** `internal/api/alerts.go`
- **Dependencies:** 13

### 27. Implement `alerts` CLI commands — `internal/cli/alerts/alerts.go`
- [ ] **Status:** Not started
- **Description:** Register `alerts` domain with subcommands: `list` (flags: `--status`, `--source`, `--since`, `--limit`, `--after`, `--full`), `get <id>`, `create` (flags: `--source-id`, `--title`, `--description`), `incidents` (list alerts attached to incidents). Compact projection on `list` by default. Add `llm-help`. Uses timeparse for `--since`.
- **Files:** `internal/cli/alerts/alerts.go`, `internal/cli/alerts/usage.go`
- **Dependencies:** 14, 15, 18, 26

### 28. Add tests for `alerts` commands
- [ ] **Status:** Not started
- **Description:** Test list, get, create, incidents subcommand. Verify API paths and query params.
- **Files:** `internal/cli/alerts/alerts_test.go`
- **Dependencies:** 27, 17

### 29. Implement `severities` API methods — `internal/api/severities.go`
- [ ] **Status:** Not started
- **Description:** Types: `Severity`. API methods: `ListSeverities(ctx)` using `GET /v1/severities`, `GetSeverity(ctx, id)` using `GET /v1/severities/{id}`.
- **Files:** `internal/api/severities.go`
- **Dependencies:** 13

### 30. Implement `severities` CLI commands — `internal/cli/severities/severities.go`
- [ ] **Status:** Not started
- **Description:** Register `severities` domain with subcommands: `list`, `get <id>`. Read-only. Add `llm-help`.
- **Files:** `internal/cli/severities/severities.go`, `internal/cli/severities/usage.go`
- **Dependencies:** 14, 15, 29

### 31. Implement `statuses` API methods — `internal/api/statuses.go`
- [ ] **Status:** Not started
- **Description:** Types: `IncidentStatus`. API methods: `ListIncidentStatuses(ctx)` using `GET /v1/incident_statuses`, `GetIncidentStatus(ctx, id)` using `GET /v1/incident_statuses/{id}`.
- **Files:** `internal/api/statuses.go`
- **Dependencies:** 13

### 32. Implement `statuses` CLI commands — `internal/cli/statuses/statuses.go`
- [ ] **Status:** Not started
- **Description:** Register `statuses` domain with subcommands: `list`, `get <id>`. Read-only. Add `llm-help`.
- **Files:** `internal/cli/statuses/statuses.go`, `internal/cli/statuses/usage.go`
- **Dependencies:** 14, 15, 31

### 33. Add tests for `severities` and `statuses` commands
- [ ] **Status:** Not started
- **Description:** Test list and get for both domains. These are simple read-only lookups.
- **Files:** `internal/cli/severities/severities_test.go`, `internal/cli/statuses/statuses_test.go`
- **Dependencies:** 30, 32, 17

### 34. Wire Phase 1 commands into root and update `llm-help`
- [ ] **Status:** Not started
- **Description:** Import and register `auth`, `incidents`, `alerts`, `severities`, `statuses` in `internal/cli/root.go`. Update `internal/cli/usage.go` llm-help text to cover all Phase 1 commands.
- **Files:** `internal/cli/root.go`, `internal/cli/usage.go`
- **Dependencies:** 20, 24, 27, 30, 32, 16

---

## Phase 2: People & Escalation

### 35. Implement `users` API methods — `internal/api/users.go`
- [ ] **Status:** Not started
- **Description:** Types: `User`, `UserCompact`. API methods: `ListUsers(ctx, query, limit, cursor)` using `GET /v2/users`, `GetUser(ctx, id)` using `GET /v2/users/{id}`.
- **Files:** `internal/api/users.go`
- **Dependencies:** 13

### 36. Implement `users` CLI commands — `internal/cli/users/users.go`
- [ ] **Status:** Not started
- **Description:** Register `users` domain with subcommands: `list` (flags: `--query`, `--limit`, `--after`, `--full`), `get <id>`. Compact projection on `list` by default (id, name, email, role). Add `llm-help`.
- **Files:** `internal/cli/users/users.go`, `internal/cli/users/usage.go`
- **Dependencies:** 14, 15, 35

### 37. Implement `roles` API methods — `internal/api/roles.go`
- [ ] **Status:** Not started
- **Description:** Types: `IncidentRole`. API methods: `ListIncidentRoles(ctx)` using `GET /v2/incident_roles`, `GetIncidentRole(ctx, id)` using `GET /v2/incident_roles/{id}`. Note: v1 also has roles endpoint; use v2.
- **Files:** `internal/api/roles.go`
- **Dependencies:** 13

### 38. Implement `roles` CLI commands — `internal/cli/roles/roles.go`
- [ ] **Status:** Not started
- **Description:** Register `roles` domain with subcommands: `list`, `get <id>`. Read-only. Add `llm-help`.
- **Files:** `internal/cli/roles/roles.go`, `internal/cli/roles/usage.go`
- **Dependencies:** 14, 15, 37

### 39. Implement `schedules` API methods — `internal/api/schedules.go`
- [ ] **Status:** Not started
- **Description:** Types: `Schedule`, `ScheduleEntry`, `ScheduleOverride`. API methods: `ListSchedules(ctx)` using `GET /v2/schedules`, `GetSchedule(ctx, id)` using `GET /v2/schedules/{id}`, `ListScheduleEntries(ctx, scheduleID, from, to)` using `GET /v2/schedule_entries` with `schedule_id` filter, `CreateScheduleOverride(ctx, params)` using `POST /v2/schedule_overrides`.
- **Files:** `internal/api/schedules.go`
- **Dependencies:** 13

### 40. Implement `schedules` CLI commands — `internal/cli/schedules/schedules.go`
- [ ] **Status:** Not started
- **Description:** Register `schedules` domain with subcommands: `list`, `get <id>`, `entries <id>` (flags: `--from`, `--to`), `override <id>` (flags: `--user`, `--from`, `--to`). Uses timeparse for `--from`/`--to`. Add `llm-help`.
- **Files:** `internal/cli/schedules/schedules.go`, `internal/cli/schedules/usage.go`
- **Dependencies:** 14, 15, 18, 39

### 41. Implement `escalations` API methods — `internal/api/escalations.go`
- [ ] **Status:** Not started
- **Description:** Types: `Escalation`, `EscalationPath`. API methods: `ListEscalations(ctx)` using `GET /v2/escalations`, `GetEscalation(ctx, id)` using `GET /v2/escalations/{id}`, `CreateEscalation(ctx, params)` using `POST /v2/escalations`, `ListEscalationPaths(ctx)` using `GET /v2/escalation_paths`, `GetEscalationPath(ctx, id)` using `GET /v2/escalation_paths/{id}`.
- **Files:** `internal/api/escalations.go`
- **Dependencies:** 13

### 42. Implement `escalations` CLI commands — `internal/cli/escalations/escalations.go`
- [ ] **Status:** Not started
- **Description:** Register `escalations` domain with subcommands: `list`, `get <id>`, `create` (flags: `--incident`, `--path`), `paths list`, `paths get <id>`. The `paths` sub-group uses a nested cobra command. Add `llm-help`.
- **Files:** `internal/cli/escalations/escalations.go`, `internal/cli/escalations/usage.go`
- **Dependencies:** 14, 15, 41

### 43. Add tests for Phase 2 commands
- [ ] **Status:** Not started
- **Description:** Tests for users (list with query, get), roles (list, get), schedules (list, get, entries with time range, override creation), escalations (list, get, create, paths list/get).
- **Files:** `internal/cli/users/users_test.go`, `internal/cli/roles/roles_test.go`, `internal/cli/schedules/schedules_test.go`, `internal/cli/escalations/escalations_test.go`
- **Dependencies:** 36, 38, 40, 42, 17

### 44. Wire Phase 2 commands into root and update `llm-help`
- [ ] **Status:** Not started
- **Description:** Register `users`, `roles`, `schedules`, `escalations` in root. Update llm-help text.
- **Files:** `internal/cli/root.go`, `internal/cli/usage.go`
- **Dependencies:** 36, 38, 40, 42, 34

---

## Phase 3: Context & Follow-through

### 45. Implement `actions` API methods — `internal/api/actions.go`
- [ ] **Status:** Not started
- **Description:** Types: `Action`. API methods: `ListActions(ctx, incidentID, limit, cursor)` using `GET /v2/actions`, `GetAction(ctx, id)` using `GET /v2/actions/{id}`. Support filtering by `incident_id` query param.
- **Files:** `internal/api/actions.go`
- **Dependencies:** 13

### 46. Implement `actions` CLI commands — `internal/cli/actions/actions.go`
- [ ] **Status:** Not started
- **Description:** Register `actions` domain with subcommands: `list` (flags: `--incident`, `--limit`, `--after`), `get <id>`. Add `llm-help`.
- **Files:** `internal/cli/actions/actions.go`, `internal/cli/actions/usage.go`
- **Dependencies:** 14, 15, 45

### 47. Implement `follow-ups` API methods — `internal/api/followups.go`
- [ ] **Status:** Not started
- **Description:** Types: `FollowUp`. API methods: `ListFollowUps(ctx, incidentID, limit, cursor)` using `GET /v2/follow_ups`, `GetFollowUp(ctx, id)` using `GET /v2/follow_ups/{id}`. Support filtering by `incident_id` query param.
- **Files:** `internal/api/followups.go`
- **Dependencies:** 13

### 48. Implement `follow-ups` CLI commands — `internal/cli/followups/followups.go`
- [ ] **Status:** Not started
- **Description:** Register `follow-ups` domain with subcommands: `list` (flags: `--incident`, `--limit`, `--after`), `get <id>`. Add `llm-help`.
- **Files:** `internal/cli/followups/followups.go`, `internal/cli/followups/usage.go`
- **Dependencies:** 14, 15, 47

### 49. Implement `catalog` API methods — `internal/api/catalog.go`
- [ ] **Status:** Not started
- **Description:** Types: `CatalogType`, `CatalogEntry`. API methods: `ListCatalogTypes(ctx)` using `GET /v2/catalog_types`, `GetCatalogType(ctx, id)` using `GET /v2/catalog_types/{id}`, `ListCatalogEntries(ctx, typeID, query, limit, cursor)` using `GET /v2/catalog_entries`, `GetCatalogEntry(ctx, id)` using `GET /v2/catalog_entries/{id}`. Note: v3 catalog endpoints also exist but v2 is sufficient.
- **Files:** `internal/api/catalog.go`
- **Dependencies:** 13

### 50. Implement `catalog` CLI commands — `internal/cli/catalog/catalog.go`
- [ ] **Status:** Not started
- **Description:** Register `catalog` domain with nested sub-groups: `types list`, `types get <id>`, `entries list` (flags: `--type`, `--query`, `--limit`, `--after`), `entries get <id>`. Add `llm-help`.
- **Files:** `internal/cli/catalog/catalog.go`, `internal/cli/catalog/usage.go`
- **Dependencies:** 14, 15, 49

### 51. Implement `custom-fields` API methods — `internal/api/customfields.go`
- [ ] **Status:** Not started
- **Description:** Types: `CustomField`. API methods: `ListCustomFields(ctx)` using `GET /v2/custom_fields`, `GetCustomField(ctx, id)` using `GET /v2/custom_fields/{id}`.
- **Files:** `internal/api/customfields.go`
- **Dependencies:** 13

### 52. Implement `custom-fields` CLI commands — `internal/cli/customfields/customfields.go`
- [ ] **Status:** Not started
- **Description:** Register `custom-fields` domain with subcommands: `list`, `get <id>`. Read-only. Add `llm-help`.
- **Files:** `internal/cli/customfields/customfields.go`, `internal/cli/customfields/usage.go`
- **Dependencies:** 14, 15, 51

### 53. Add tests for Phase 3 commands
- [ ] **Status:** Not started
- **Description:** Tests for actions (list with incident filter, get), follow-ups (list with incident filter, get), catalog (types list/get, entries list with type filter/get), custom-fields (list, get).
- **Files:** `internal/cli/actions/actions_test.go`, `internal/cli/followups/followups_test.go`, `internal/cli/catalog/catalog_test.go`, `internal/cli/customfields/customfields_test.go`
- **Dependencies:** 46, 48, 50, 52, 17

### 54. Wire Phase 3 commands into root and update `llm-help`
- [ ] **Status:** Not started
- **Description:** Register `actions`, `follow-ups`, `catalog`, `custom-fields` in root. Update llm-help text.
- **Files:** `internal/cli/root.go`, `internal/cli/usage.go`
- **Dependencies:** 46, 48, 50, 52, 44

---

## Phase 4: Communication

### 55. Implement `status-pages` API methods — `internal/api/statuspages.go`
- [ ] **Status:** Not started
- **Description:** Types: `StatusPage`, `StatusPageIncident`, `StatusPageIncidentUpdate`. API methods: `ListStatusPages(ctx)` using `GET /v2/status_pages`, `ListStatusPageIncidents(ctx, pageID)` using `GET /v2/status_page_incidents`, `CreateStatusPageIncident(ctx, params)` using `POST /v2/status_page_incidents`, `UpdateStatusPageIncident(ctx, id, params)` using `PUT /v2/status_page_incidents/{id}`, `CreateStatusPageIncidentUpdate(ctx, params)` using `POST /v2/status_page_incident_updates`.
- **Files:** `internal/api/statuspages.go`
- **Dependencies:** 13

### 56. Implement `status-pages` CLI commands — `internal/cli/statuspages/statuspages.go`
- [ ] **Status:** Not started
- **Description:** Register `status-pages` domain with subcommands: `list`, `incidents list` (flags: `--page`), `incidents create` (flags: `--page`, `--name`), `incidents update <id>` (flags: `--status`). Nested `incidents` sub-group. Add `llm-help`.
- **Files:** `internal/cli/statuspages/statuspages.go`, `internal/cli/statuspages/usage.go`
- **Dependencies:** 14, 15, 55

### 57. Add tests for Phase 4 commands
- [ ] **Status:** Not started
- **Description:** Tests for status-pages (list, incidents list/create/update).
- **Files:** `internal/cli/statuspages/statuspages_test.go`
- **Dependencies:** 56, 17

### 58. Wire Phase 4 commands into root and update `llm-help`
- [ ] **Status:** Not started
- **Description:** Register `status-pages` in root. Final llm-help update with all domains.
- **Files:** `internal/cli/root.go`, `internal/cli/usage.go`
- **Dependencies:** 56, 54

---

## Phase 5: Polish & Distribution

### 59. Create Claude Code skill for agent-incident
- [ ] **Status:** Not started
- **Description:** Create `skills/agent-incident/SKILL.md` with tool description, usage examples, and domain reference. Enables other Claude Code agents to discover and use this tool. Match `skills/agent-dd/SKILL.md` pattern.
- **Files:** `skills/agent-incident/SKILL.md`
- **Dependencies:** 58

### 60. End-to-end smoke test with real API
- [ ] **Status:** Not started
- **Description:** Manual verification checklist (not automated): `auth add`, `auth check`, `incidents list`, `incidents get`, `alerts list`, `severities list`, `statuses list`. Verify JSON/YAML/NDJSON output modes. Verify `--full` flag. Verify error output for bad auth.
- **Dependencies:** 34

### 61. First release
- [ ] **Status:** Not started
- **Description:** Tag `v0.1.0`, run `goreleaser release`. Verify binaries for all platforms. Update homebrew tap if applicable.
- **Dependencies:** 60, 5

---

## Review Notes (2026-04-09)

Changes made during review:

1. **Task 3 (root.go):** Added `allGlobals()` helper pattern and noted that auth.Register takes only `root` (no globals func) while domain commands take `globals func() *shared.GlobalFlags` — matches agent-dd convention where `org.Register(root)` differs from `monitors.Register(root, allGlobals)`.

2. **Task 5 (.goreleaser.yml):** Added missing details: windows/arm64 skip, zip format override for windows, changelog section with sort/exclude filters — all present in agent-dd's goreleaser config but omitted here.

3. **Task 7:** Added `AGENTS.md` alongside `CLAUDE.md`. Both sibling projects (agent-dd, agent-statsig) have this file.

4. **Task 14 (shared.go):** Added `GlobalsFunc` type alias (present in agent-dd). Clarified default output formats: `WritePaginatedList` defaults to NDJSON, `WriteItem` defaults to JSON — matching the design doc's output format table.

5. **Task 27 (alerts CLI):** Added explicit mention of compact projection on `list` by default. The task had `--full` flag but didn't state that list output is compact without it.

6. **Task 36 (users CLI):** Added `--full` flag and compact projection note. Design doc specifies compact projections for all list commands, but users list was missing the `--full` flag entirely.

### Items reviewed and found correct (no changes needed):

- **Auth model completeness:** All 5 auth subcommands (add, check, default, list, remove) plus `--organization` global flag are covered.
- **Output format:** NDJSON/JSON/YAML with correct defaults, null pruning, and `--format` flag all covered in tasks 9 and 14.
- **Dependency chains:** No circular dependencies found. All chains are correct.
- **Phase ordering:** Phases are well-structured. Infrastructure (0/0.5) before domains (1-4) before polish (5).
- **Design doc coverage:** All 13 proposed domains + version + auth are represented. All skipped domains are correctly excluded.
- **File paths and package names:** Match sibling conventions (`internal/cli/<domain>/`, `internal/api/<domain>.go`).
- **Compact vs full:** Properly handled for incidents; now also explicit for alerts and users. Severities, statuses, roles, and custom-fields are small reference types where compact mode adds no value — correctly omitted.

### Items NOT added (considered but rejected):

- **README.md / LICENSE:** Both siblings have these but they're boilerplate, not implementation tasks. Can be added at release time without blocking anything.
- **Splitting large tasks:** Tasks like 19 (shared infra tests) and 53 (Phase 3 tests) bundle multiple test files, but they're naturally parallelizable within a single task and splitting them would add coordination overhead without value.
- **CI config (.github/workflows):** Neither sibling appears to have checked-in CI config (goreleaser handles release). Not adding.
- **Mock server binary (agent-dd has `cmd/mockdd`):** agent-incident uses `shared.ClientFactory` + `httptest.Server` for testing, which is simpler and sufficient. No separate mock binary needed.
- **API-level test files (e.g., `internal/api/incidents_test.go`):** agent-dd has these but task 19 already covers `client_test.go` for HTTP error classification and core client logic. Domain-specific API method tests are implicitly covered by the CLI command tests which exercise the full stack through the mock server.
