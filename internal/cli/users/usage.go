package users

const llmHelpText = `agent-incident users — incident.io user management

COMMANDS
  agent-incident users list               List users (compact view by default)
  agent-incident users get <id>           Get a single user by ID

FLAGS (list)
  --query <text>    Filter users by name or email
  --limit <n>       Maximum number of users to return
  --after <cursor>  Pagination cursor for the next page
  --full            Show all fields instead of compact view

COMPACT FIELDS (default)
  id       Unique user identifier
  name     Display name
  email    Email address
  role     User role (e.g. "owner", "admin", "responder", "viewer")

FULL FIELDS (with --full)
  id             Unique user identifier
  name           Display name
  email          Email address
  role           User role
  slack_user_id  Linked Slack user ID
  created_at     ISO 8601 creation timestamp
  updated_at     ISO 8601 last-update timestamp

NOTES
  Users are read-only — they are managed in incident.io or synced from your
  identity provider. Use the query flag to search by name or email.

EXAMPLES
  # List all users (compact)
  agent-incident users list

  # Search for a user by name
  agent-incident users list --query "Alice"

  # Get full details for a specific user
  agent-incident users get USR123

  # Paginate through users
  agent-incident users list --limit 25 --after <cursor>
`
