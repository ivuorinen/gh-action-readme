#!/usr/bin/env bash
# PostToolUse: run yamlfmt on edited YAML files.
# Uses go run like the Makefile — no global install required, version pinned via Makefile.
# testdata/ excluded to match the pre-commit yamlfmt exclusion pattern.
set -uo pipefail

input=$(cat)
file_path=$(echo "$input" | python3 -c \
  "import sys,json; d=json.load(sys.stdin); print(d.get('tool_input',{}).get('file_path',''))" \
  2>/dev/null || echo "")

[[ "$file_path" == *.yml || "$file_path" == *.yaml ]] || exit 0
[[ "$file_path" == */testdata/* ]] && exit 0

REPO_ROOT=$(git -C "$(dirname "$file_path")" rev-parse --show-toplevel 2>/dev/null) || exit 0
cd "$REPO_ROOT" || {
  echo "Error: cannot cd to repo root: $REPO_ROOT" >&2
  exit 1
}

YAMLFMT_MODULE="github.com/google/yamlfmt/cmd/yamlfmt"
YAMLFMT_VERSION=$(awk '/^YAMLFMT_VERSION :=/ { print $3; exit }' "$REPO_ROOT/Makefile" 2>/dev/null)
if [[ -z "$YAMLFMT_VERSION" ]]; then
  echo "Warning: YAMLFMT_VERSION not found in Makefile — skipping format"
  exit 0
fi

echo "→ yamlfmt ${YAMLFMT_VERSION} $(basename "$file_path")"
go run "${YAMLFMT_MODULE}@${YAMLFMT_VERSION}" "$file_path" 2>&1
exit 0
