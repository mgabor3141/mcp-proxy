# Call hooks

Sample `callHook.command` scripts. Each reads the tool-call request as JSON on
stdin and signals the decision via **exit code**: `0` approves, non-zero denies.
Anything printed to stderr/stdout becomes the denial reason the model sees.

Available env vars: `MCP_SERVER`, `MCP_TOOL`.

Wire one up in your config:

```json
{
  "mcpServers": {
    "github": {
      "command": "github-mcp-server",
      "args": ["stdio"],
      "options": {
        "callHook": {
          "command": ["examples/call-hooks/approve-editor.sh"],
          "requireFor": ["create_issue", "merge_pull_request"],
          "timeoutSec": 120
        }
      }
    }
  }
}
```

- `approve-editor.sh` — opens the request in `$EDITOR`; approve by leaving a line
  starting with `APPROVE`. Host-side (needs a terminal/editor).
- `approve-notify.sh` — desktop notification with the tool name; approves only
  if a confirmation file is touched. Host-side (needs a desktop session).
- `deny.sh` — always denies (useful for testing / hard blocks).
