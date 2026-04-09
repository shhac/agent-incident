package schedules

const schedulesLLMHelp = `# schedules domain — agent-incident CLI

## Commands

### schedules list
List all on-call schedules.

### schedules get <name-or-id>
Retrieve a single schedule by name or ID, including timezone and metadata.

### schedules entries <name-or-id>
List schedule entries (who is on-call) for a time window.
  --from  Window start (supports relative: now-1h, RFC3339, unix epoch; default: now-1h)
  --to    Window end (supports relative: now+1h, RFC3339, unix epoch; default: now)

### schedules override <name-or-id>
Create an override on a schedule (temporarily replace who is on-call).
  --user  User name, email, or ID for the override (required)
  --from  Override start time (required; supports relative, RFC3339, epoch)
  --to    Override end time (required; supports relative, RFC3339, epoch)

## Name Resolution
Schedule commands accept a name or ID. If the value doesn't look like a ULID,
it is matched against schedule names (case-insensitive, substring match).
The --user flag on override also resolves by name or email.

## Common Workflows
- See who is on-call now: schedules entries Engineering
- See on-call for next 24h: schedules entries "Primary On-Call" --from now --to now+24h
- Cover for someone: schedules override Engineering --user alice@example.com --from now --to now+4h
- List all schedules: schedules list
`
