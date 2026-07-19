#!/usr/bin/env bash
# PreToolUse(Bash): block prohibited / destructive shell commands.
# Enforces:
#   - git-conduct.md        — no --no-verify/-n, no force-push (incl. +refspec),
#                             no core.hooksPath override.
#   - no-commit-bylines.md  — no Co-Authored-By/Signed-off-by attribution trailers.
#   - readme-protection.md  — no Bash write to the repo-root README.md.
#   - pre-sensitive-file-guard.sh — no Bash write to credential files.
#   - filesystem destruction — no rm -rf / find -delete on root/home/cwd.
# Exit 1 blocks the tool call.
set -uo pipefail

input=$(cat)
cmd=$(echo "$input" | python3 -c \
  "import sys,json; d=json.load(sys.stdin); print(d.get('tool_input',{}).get('command',''))" \
  2>/dev/null || echo "")

[[ -z "$cmd" ]] && exit 0

# Normalize newlines so multiline commands (git commit \\\n --no-verify) match.
cmd_flat=$(echo "$cmd" | tr '\n' ' ')
# Lowercased copy for case-insensitive filename/trailer matching (macOS APFS).
cmd_lower=$(printf '%s' "$cmd_flat" | tr '[:upper:]' '[:lower:]')

block() {
  echo "BLOCKED: $1"
  exit 1
}

# --- git-conduct.md ------------------------------------------------------------

# --no-verify (git commit|push) or the -n short flag / cluster (git commit -n,
# -nm, -mn) bypasses required hooks. -n on push is --dry-run, so the short-flag
# cluster is only rejected for commit; --no-verify is rejected for both.
if echo "$cmd_flat" | grep -qE "git[[:space:]]+(commit|push)([[:space:]]|$)" 2>/dev/null &&
  echo "$cmd_flat" | grep -qE "(^|[[:space:]])--no-ver[a-z-]*([[:space:]]|=|$)" 2>/dev/null; then
  block "--no-verify skips required git hooks (git-conduct.md). Fix the underlying hook failure instead."
fi
if echo "$cmd_flat" | grep -qE "git[[:space:]]+commit([[:space:]]|$)" 2>/dev/null &&
  echo "$cmd_flat" | grep -qE "(^|[[:space:]])-[a-zA-Z]*n[a-zA-Z]*([[:space:]]|$)" 2>/dev/null; then
  block "git commit -n/-nm skips required hooks (git-conduct.md). Fix the underlying hook failure instead."
fi

# core.hooksPath override points git at an alternate/empty hooks dir, either via
# the `git -c` token or the equivalent GIT_CONFIG_* / GIT_CONFIG_GLOBAL|SYSTEM
# environment injection (no `git -c` present).
if echo "$cmd_lower" | grep -qE "git[[:space:]]+-c[[:space:]]+core\.hookspath" 2>/dev/null ||
  echo "$cmd_lower" | grep -qE "(git_config_key_[0-9]+=core\.hookspath|git_config_(global|system)=)" 2>/dev/null; then
  block "overriding core.hooksPath (via git -c or GIT_CONFIG_* env) bypasses required git hooks (git-conduct.md)."
fi

# Force-push: --force, -f, --force-with-lease, or a +refspec (git push origin +main).
if echo "$cmd_flat" | grep -qE "git[[:space:]]+push([[:space:]]|$)" 2>/dev/null &&
  echo "$cmd_flat" | grep -qE "(^|[[:space:]])(--for[a-z-]*|-[a-zA-Z]*f[a-zA-Z]*|\+[^[:space:]]+)([[:space:]]|$)" 2>/dev/null; then
  block "force-push is prohibited (git-conduct.md). Use a new branch or a revert commit."
fi

# --- no-commit-bylines.md ------------------------------------------------------

if echo "$cmd_flat" | grep -qE "git[[:space:]]+commit([[:space:]]|$)" 2>/dev/null &&
  echo "$cmd_lower" | grep -qE "(co-authored-by|signed-off-by|co-committed-by)[[:space:]]*:" 2>/dev/null; then
  block "commit attribution trailers are prohibited (no-commit-bylines.md). Remove Co-Authored-By/Signed-off-by lines."
fi
# -s / --signoff (incl. short-flag clusters like -sm) makes git GENERATE a
# Signed-off-by trailer with no literal trailer text in the command.
if echo "$cmd_flat" | grep -qE "git[[:space:]]+commit([[:space:]]|$)" 2>/dev/null &&
  echo "$cmd_flat" | grep -qE "(^|[[:space:]])(--signoff|-[a-zA-Z]*s[a-zA-Z]*)([[:space:]]|$)" 2>/dev/null; then
  block "-s/--signoff generates a Signed-off-by trailer (no-commit-bylines.md). Remove it."
fi

# --- readme-protection.md ------------------------------------------------------

# Only scrutinize commands that mention README.md; block the write forms.
if echo "$cmd_lower" | grep -qE "readme\.md" 2>/dev/null &&
  echo "$cmd_lower" | grep -qE \
    "(>>?[[:space:]]*[^|&;]*readme\.md|(cp|mv|tee|install|sed[[:space:]]+-i|dd|ln|truncate)[[:space:]][^|&;]*readme\.md|git[[:space:]]+(checkout|restore)[[:space:]][^|&;]*readme\.md)" \
    2>/dev/null; then
  block "README.md is protected (readme-protection.md). Do not write it from Bash; use /tmp/ or testdata/."
fi

# --- sensitive-file write guard (Bash counterpart) -----------------------------

sens='(\.env(\.[a-z0-9]+)?|\.envrc|\.netrc|\.pgpass|id_rsa|id_dsa|id_ecdsa|id_ed25519'
sens+='|[^[:space:]<>|&;]*\.(pem|key|p12|pfx|p8|der)'
sens+='|[^[:space:]<>|&;]*(secret|credential|password)[^[:space:]<>|&;]*)'
if echo "$cmd_lower" | grep -qE \
  "(>>?[[:space:]]*[^|&;]*${sens}([[:space:]\"']|$)|(cp|mv|tee|install)[[:space:]][^|&;]*${sens}([[:space:]\"']|$))" \
  2>/dev/null; then
  block "writing a credential/secret file from Bash is blocked (pre-sensitive-file-guard.sh). Edit manually."
fi

# --- filesystem destruction ----------------------------------------------------

# rm with a recursive/force flag targeting root, home, cwd, or a system dir.
rm_prefix='(^|[[:space:]|;&(])rm[[:space:]]+(-[a-z]*[rf][a-z]*[[:space:]]+)+'
rm_target='('
rm_target+='/([[:space:]"*]|$)'                 # / or /*
rm_target+='|~([[:space:]/"]|$)'                # ~
rm_target+='|"?\$\{?home\}?"?([[:space:]/"]|$)' # $HOME / ${HOME} / "$HOME"
rm_target+='|\.\.?([[:space:]"]|$)'             # . or ..
rm_target+='|/(etc|usr|bin|sbin|var|lib|system|library|opt|boot|dev|root)([[:space:]/"]|$)'
rm_target+='|/users/[a-z0-9._-]+([[:space:]"]|$)' # home root, not a subpath
rm_target+=')'
if echo "$cmd_lower" | grep -qE "${rm_prefix}${rm_target}" 2>/dev/null; then
  block "destructive rm on a root/home/cwd/system path is not allowed."
fi

# find rooted at /, ~, $HOME, or . combined with -delete / -exec rm.
if echo "$cmd_lower" | grep -qE \
  "find[[:space:]]+(\"?/\"?|~|\"?\$\{?home\}?\"?|\.)[[:space:]].*(-delete|-exec[[:space:]]+rm)" \
  2>/dev/null; then
  block "find ... -delete/-exec rm on a root/home path is not allowed."
fi

exit 0
