#!/usr/bin/env bash
# PostToolUse: run golangci-lint --fix on the package containing the edited Go file.
# Uses go run like the Makefile — no global install required, version pinned via Makefile.
set -uo pipefail

input=$(cat)
file_path=$(echo "$input" | python3 -c \
  "import sys,json; d=json.load(sys.stdin); print(d.get('tool_input',{}).get('file_path',''))" \
  2>/dev/null || echo "")

[[ "$file_path" == *.go ]] || exit 0

REPO_ROOT=$(git -C "$(dirname "$file_path")" rev-parse --show-toplevel 2>/dev/null) || exit 0
cd "$REPO_ROOT"

GOLANGCI_MODULE="github.com/golangci/golangci-lint/v2/cmd/golangci-lint"
GOLANGCI_VERSION=$(awk '/^GOLANGCI_LINT_VERSION :=/ { print $3; exit }' "$REPO_ROOT/Makefile" 2>/dev/null)
if [[ -z "$GOLANGCI_VERSION" ]]; then
  echo "Warning: GOLANGCI_LINT_VERSION not found in Makefile — skipping lint"
  exit 0
fi

rel_dir=$(dirname "${file_path#${REPO_ROOT}/}")
if [[ "$rel_dir" == "." ]]; then
  pkg="./..."
else
  pkg="./${rel_dir}/..."
fi

echo "→ golangci-lint ${GOLANGCI_VERSION} --fix ${pkg}"

rc=0
output=$(go run "${GOLANGCI_MODULE}@${GOLANGCI_VERSION}" run --fix "$pkg" 2>&1) || rc=$?

if [[ -n "${output:-}" ]]; then
  echo "$output" | head -60
fi

exit $rc
