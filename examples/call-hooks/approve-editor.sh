#!/usr/bin/env bash
# Approve a gated MCP tool call by reviewing it in $EDITOR.
# Approve by leaving a line beginning with APPROVE; otherwise the call is denied.
set -euo pipefail
req="$(cat)"   # tool-call request JSON on stdin
tmp="$(mktemp --suffix=.json)"
trap 'rm -f "$tmp"' EXIT
{
  echo "# Reviewing ${MCP_SERVER:-?}/${MCP_TOOL:-?}"
  echo "# To APPROVE: keep a line starting with APPROVE. Save & quit to decide."
  echo "APPROVE"
  echo "# --- request ---"
  echo "$req" | (command -v jq >/dev/null && jq . || cat)
} > "$tmp"
"${EDITOR:-vi}" "$tmp" >/dev/tty 2>/dev/tty </dev/tty
if grep -q '^APPROVE' "$tmp"; then exit 0; fi
echo "operator did not approve in editor" >&2
exit 1
