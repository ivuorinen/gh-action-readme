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

# Case-insensitive compare: on macOS APFS, README.md/readme.md/ReadMe.md are the
# same file, so lowercase both sides before matching.
abs_lower=$(printf '%s' "$abs_file_path" | tr '[:upper:]' '[:lower:]')
target_lower=$(printf '%s' "${REPO_ROOT}/README.md" | tr '[:upper:]' '[:lower:]')

if [[ "$abs_lower" == "$target_lower" ]]; then
  echo "BLOCKED: /README.md is protected (readme-protection.md). Use /tmp/ or testdata/ for test output."
  exit 1
fi

exit 0
