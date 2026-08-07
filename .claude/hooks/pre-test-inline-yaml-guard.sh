#!/usr/bin/env bash
# PreToolUse(Edit|Write): enforce no-inline-yaml-in-tests.md.
# Blocks edits to *_test.go whose pending content embeds a YAML-shaped document in
# a Go string literal (raw backtick, or double-quoted with \n escapes decoded).
#
# The detection lives in inline_yaml_scan.py so the identical logic also runs as a
# pre-commit hook over staged files — a PreToolUse hook is advisory by construction
# and cannot be the only line of defence.
# Exit 1 blocks the tool call.
set -uo pipefail

hook_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)

if ! command -v python3 >/dev/null 2>&1; then
  # Fail open: a missing interpreter must not wedge every edit.
  exit 0
fi

python3 "${hook_dir}/inline_yaml_scan.py" --stdin-content
exit $?
