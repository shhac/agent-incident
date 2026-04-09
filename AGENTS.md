# agent-incident: Subagent Roles

## Implementer

Writes production code for a specific domain (API methods + CLI commands). Owns files in `internal/api/<domain>.go` and `internal/cli/<domain>/`. Follows existing patterns from sibling domains.

## Reviewer

Reviews implementation for correctness, consistency with sibling patterns, and test coverage. Does not modify files — reports findings only.

## Test Writer

Writes tests for a specific domain using `shared.SetupMockServer`. Owns `*_test.go` files. Verifies API paths, query params, request bodies, and output format.
