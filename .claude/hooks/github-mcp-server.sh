#!/usr/bin/env bash
# GitHub MCP server wrapper.
# Validates token presence before launching Docker so failures surface at startup
# with a clear message, not as cryptic API errors during tool calls.
set -euo pipefail

TOKEN_VAR="GITHUB_PERSONAL_ACCESS_TOKEN"
if [[ -z "${GITHUB_PERSONAL_ACCESS_TOKEN:-}" ]]; then
  echo "Error: ${TOKEN_VAR} is not set." >&2
  echo "Set it in your shell profile: export ${TOKEN_VAR}=ghp_..." >&2
  exit 1
fi

exec docker run -i --rm \
  -e GITHUB_PERSONAL_ACCESS_TOKEN \
  ghcr.io/github/github-mcp-server "$@"
