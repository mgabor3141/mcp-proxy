#!/usr/bin/env bash
# Desktop-notification approval. Sends a notification and waits for the operator
# to create the ack file (e.g. via the notification action or manually).
set -euo pipefail
cat >/dev/null   # drain stdin (request JSON); not used here
ack="/tmp/mcp-approve-${MCP_SERVER:-x}-${MCP_TOOL:-x}"
rm -f "$ack"
msg="Approve ${MCP_SERVER:-?}/${MCP_TOOL:-?}? touch $ack"
if command -v notify-send >/dev/null; then notify-send "MCP approval" "$msg"
elif command -v terminal-notifier >/dev/null; then terminal-notifier -title "MCP approval" -message "$msg"
else echo "$msg" >&2; fi
for _ in $(seq 1 "${MCP_APPROVE_WAIT:-60}"); do
  [ -e "$ack" ] && { rm -f "$ack"; exit 0; }
  sleep 1
done
echo "no approval ack within timeout" >&2
exit 1
