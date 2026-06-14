# Configuration

This project supports a v2 JSON configuration. v1 configs are automatically migrated at load time.

- Online converter (build Claude config from your proxy): https://tbxark.github.io/mcp-proxy

## Full Example

```jsonc
{
  "mcpProxy": {
    "baseURL": "https://mcp.example.com",
    "addr": ":9090",
    "name": "MCP Proxy",
    "version": "1.0.0",
    "type": "streamable-http", // or "sse" (default)
    "options": {
      "panicIfInvalid": false,
      "logEnabled": true,
      "authTokens": ["DefaultToken"]
    }
  },
  "mcpServers": {
    "github": {
      // stdio client
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": { "GITHUB_PERSONAL_ACCESS_TOKEN": "<YOUR_TOKEN>" },
      "options": {
        "toolFilter": {
          "mode": "block",
          "list": ["create_or_update_file"]
        }
      }
    },
    "fetch": {
      // stdio client
      "command": "uvx",
      "args": ["mcp-server-fetch"],
      "options": {
        "panicIfInvalid": true,
        "logEnabled": false,
        "authTokens": ["SpecificToken"]
      }
    },
    "amap": {
      // SSE client
      "url": "https://mcp.amap.com/sse?key=<YOUR_TOKEN>",
      "options": {
        "disabled": true
      }
    }
  }
}
```

## mcpProxy

- `baseURL`: Public URL base used to build client endpoints.
- `addr`: Bind address (e.g. `:9090`).
- `name`, `version`: Server identity for MCP handshake.
- `type`: `sse` (default) or `streamable-http`.
- `options`: Defaults inherited by `mcpServers.*.options` (can be overridden per server).

## mcpServers

Each entry defines a downstream MCP server. Supported client types:

- `stdio` (implicit when `command` is set): run a subprocess via stdio.
- `sse` (implicit when `url` is set and `transportType` ≠ `streamable-http`): connect via Server‑Sent Events.
- `streamable-http` (requires `transportType: "streamable-http"`): connect via HTTP streaming.

Common fields:

- `command`, `args`, `env` — for `stdio` clients.
- `url`, `headers` — for `sse` and `streamable-http` clients.
- `timeout` — request timeout for `streamable-http`.
- `options` — per‑server overrides and filters (see below).

## options

- `panicIfInvalid` (bool): If true, startup fails when a client cannot initialize.
- `logEnabled` (bool): Log requests and events for this client.
- `authTokens` ([]string): Valid bearer tokens; requests must include `Authorization: <token>`.
- `toolFilter` (object): Selectively expose tools to the proxy:
  - `mode`: `allow` or `block`.
  - `list`: List of tool names.
- `callHook` (object): Gate selected tools behind an external command at call time.
  Where `toolFilter` decides tool *visibility* once at startup, `callHook` runs
  *per invocation*, so it can implement human-in-the-loop approval,
  audit-with-veto, policy checks, etc.
  - `command` ([]string): Program + args to run before forwarding the call.
  - `requireFor` ([]string): Exact tool names that require approval. Tools not
    listed pass straight through (no hook overhead).
  - `timeoutSec` (int): Max seconds to wait for the hook. Default `120`.

  The command receives the tool-call request as JSON on **stdin**, plus
  `MCP_SERVER` and `MCP_TOOL` environment variables. **Exit code `0` approves**
  the call; any non-zero exit, spawn error, or timeout **denies** it
  (fail-closed). On denial the call returns an MCP `isError` result whose text
  is the hook's stderr/stdout, so the model sees a readable reason and can
  adapt. See `examples/call-hooks/` for sample scripts.
- `Disabled` (bool): Enable or disable this server. Disabled servers are skipped at startup.

Notes:

- `mcpProxy.options.authTokens` serves as the default token set if a server omits `options.authTokens`.
- `mcpProxy.options.callHook` serves as the default if a server omits `options.callHook`.
- To discover tool names for filtering, start without a filter and check logs for lines like `<server> Adding tool <name>`.

