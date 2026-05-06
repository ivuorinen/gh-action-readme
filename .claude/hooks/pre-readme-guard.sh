#!/usr/bin/env bash
# PreToolUse: enforce readme-protection.md — block any write to /README.md.
# Exit 1 blocks the tool call.
set -uo pipefail

input=$(cat)
file_path=$(echo "$input" | python3 -c \
  "import sys,json; d=json.load(sys.stdin); print(d.get('tool_input',{}).get('file_path',''))" \
  2>/dev/null || echo "")

[[ -z "$file_path" ]] && exit 0

REPO_ROOT=$(git -C "$(dirname "$file_path")" rev-parse --show-toplevel 2>/dev/null) || exit 0

# Normalize to absolute path so relative inputs like "./README.md" are caught
abs_file_path=$(cd "$(dirname "$file_path")" 2>/dev/null && echo "$PWD/$(basename "$file_path")") || abs_file_path="$file_path"

if [[ "$abs_file_path" == "${REPO_ROOT}/README.md" ]]; then
  echo "BLOCKED: /README.md is protected (readme-protection.md). Use /tmp/ or testdata/ for test output."
  exit 1
fi

exit 0
