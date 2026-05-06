#!/usr/bin/env bash
# PreToolUse: block catastrophically destructive shell commands.
# Enforces git-conduct.md (no --no-verify, no force-push) and prevents filesystem destruction.
# Exit 1 blocks the tool call.
set -uo pipefail

input=$(cat)
cmd=$(echo "$input" | python3 -c \
  "import sys,json; d=json.load(sys.stdin); print(d.get('tool_input',{}).get('command',''))" \
  2>/dev/null || echo "")

[[ -z "$cmd" ]] && exit 0

# Normalize newlines so multiline commands (git commit \\\n  --no-verify) are matched correctly.
cmd_flat=$(echo "$cmd" | tr '\n' ' ')

# git-conduct.md: --no-verify bypasses required hooks
if echo "$cmd_flat" | grep -qE "git[[:space:]]+(commit|push)[[:space:]].*--no-verify" 2>/dev/null; then
  echo "BLOCKED: --no-verify skips required git hooks (git-conduct.md). Fix the underlying hook failure instead."
  exit 1
fi

# Block any force push — bare 'git push --force' (no explicit branch) is equally destructive.
# Catches --force, -f, and --force-with-lease.
if echo "$cmd_flat" | grep -qE "git[[:space:]]+push[[:space:]]" 2>/dev/null &&
  echo "$cmd_flat" | grep -qE "(^|[[:space:]])(--force|--force-with-lease|-f)([[:space:]]|$)" 2>/dev/null; then
  echo "BLOCKED: force-push is prohibited (git-conduct.md). Use a new branch or revert commit."
  exit 1
fi

# Block rm -rf on / or ~ (filesystem destruction)
if echo "$cmd_flat" | grep -qE "rm[[:space:]]+-[rf]{1,2}[[:space:]]+(\/|~)([[:space:]]|\/|$)" 2>/dev/null; then
  echo "BLOCKED: rm -rf on root or home directory is not allowed."
  exit 1
fi

exit 0
