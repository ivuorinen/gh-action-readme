#!/usr/bin/env bash
# PreToolUse(Edit|Write): enforce no-inline-yaml-in-tests.md.
# Blocks edits to *_test.go whose new content embeds a multi-line YAML mapping in
# a Go raw-string (backtick) literal. Heuristic and low false-positive: it only
# fires on a backtick block with >= 2 YAML key lines AND at least one indented
# (nested) key line — the shape of an inlined action.yml fixture.
# Exit 1 blocks the tool call.
set -uo pipefail

input=$(cat)

CLAUDE_HOOK_INPUT="$input" python3 <<'PY'
import os, json, re, sys

try:
    d = json.loads(os.environ.get("CLAUDE_HOOK_INPUT", "") or "{}")
except (ValueError, TypeError):
    sys.exit(0)  # fail open: never block on a parse error

ti = d.get("tool_input", {}) or {}
path = ti.get("file_path", "") or ""
if not path.endswith("_test.go"):
    sys.exit(0)

# Edit provides new_string; Write provides content.
content = ti.get("new_string") or ti.get("content") or ""

key = re.compile(r"^[ \t]*[A-Za-z0-9_.-]+:(?: |$)")
indented_key = re.compile(r"^[ \t]+[A-Za-z0-9_.-]+:(?: |$)")

# Raw-string (backtick) literals, plus double-quoted literals with the \n
# escapes decoded — inline YAML as "a: b\n  c: d\n" is a common evasion of the
# backtick-only scan.
blocks = re.findall(r"`([^`]*)`", content, re.S)
for quoted in re.findall(r'"((?:[^"\\]|\\.)*)"', content):
    try:
        blocks.append(quoted.encode("latin-1", "ignore").decode("unicode_escape"))
    except (UnicodeDecodeError, UnicodeEncodeError):
        pass

for block in blocks:
    lines = block.splitlines()
    keyed = [ln for ln in lines if key.match(ln)]
    if len(keyed) >= 2 and any(indented_key.match(ln) for ln in lines):
        sys.exit(7)  # inline YAML detected

sys.exit(0)
PY
rc=$?

if [[ $rc -eq 7 ]]; then
  echo "BLOCKED: inline YAML in a *_test.go edit (no-inline-yaml-in-tests.md). Move it to testdata/yaml-fixtures/, load via testutil, and register the path in testutil/test_constants.go."
  exit 1
fi

exit 0
