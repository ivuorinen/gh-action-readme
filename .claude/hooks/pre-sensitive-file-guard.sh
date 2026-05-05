#!/usr/bin/env bash
# PreToolUse: block writes to files whose names suggest sensitive content.
# Exit 1 blocks the tool call and surfaces the message to Claude.
set -uo pipefail

input=$(cat)
file_path=$(echo "$input" | python3 -c \
  "import sys,json; d=json.load(sys.stdin); print(d.get('tool_input',{}).get('file_path',''))" \
  2>/dev/null || echo "")

[[ -z "$file_path" ]] && exit 0

basename_lower=$(basename "$file_path" | tr '[:upper:]' '[:lower:]')

# Block .env files
if [[ "$basename_lower" == ".env" || "$basename_lower" == .env.* ]]; then
  echo "BLOCKED: .env files may contain secrets. Edit manually outside Claude."
  exit 1
fi

# Block certificate and private key files by extension
if echo "$basename_lower" | grep -qE "\.(pem|key|p12|pfx|p8|der)$" 2>/dev/null; then
  echo "BLOCKED: Filename suggests a private key or certificate: $(basename "$file_path"). Edit manually."
  exit 1
fi

# Block secret-store config/data files by name using word-boundary matching.
# Code files are exempt — they handle secrets programmatically, not store them.
case "$basename_lower" in
*.go | *.ts | *.js | *.jsx | *.tsx | *.py | *.rb | *.java | *.rs | *.c | *.cpp | *.h)
  exit 0
  ;;
esac

if echo "$basename_lower" | grep -qE "\b(secret|credential|password)(s)?\b" 2>/dev/null; then
  echo "BLOCKED: Filename suggests a secret store: $(basename "$file_path"). Edit manually."
  exit 1
fi

exit 0
