#!/usr/bin/env bash
# Always deny. Useful as a hard block or for testing the deny path.
cat >/dev/null
echo "blocked by policy (deny.sh)" >&2
exit 1
