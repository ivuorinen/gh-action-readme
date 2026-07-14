#!/usr/bin/env bash
# PreToolUse(Bash): enforce .claude/rules/context-mode-required.md.
#
# Denies raw Bash commands that PRODUCE or GATHER large output for analysis
# (go test, grep/find/cat, git log/diff, make test|lint|security|coverage, and
# pipelines feeding head/tail/wc). The rule mandates routing these through the
# context-mode tool (ctx_execute / ctx_batch_execute / ctx_execute_file) so raw
# bytes stay in the sandbox and only a derived summary enters the conversation.
#
# The rule forbids the `# ctx-ok` bypass on context-producing commands, so this
# hook does NOT honor it for the denylisted verbs.
#
# State-mutation, short OBSERVE, package installs, and interactive commands are
# exempt (the must-run-direct allowlist) and pass through.
#
# Blocking mechanism: emits a PreToolUse `permissionDecision: deny` (the
# documented, unambiguous block) rather than relying on an exit code.
# Fails closed: on malformed input or an internal error it denies rather than
# silently allowing, so a parse failure never opens the gap it guards.
set -uo pipefail

deny() {
  # $1 = reason shown to the agent
  printf '{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":%s}}\n' \
    "$(printf '%s' "$1" | python3 -c 'import sys,json; print(json.dumps(sys.stdin.read()))' 2>/dev/null || printf '"context-mode routing required"')"
  exit 0
}

input=$(cat)

cmd=$(printf '%s' "$input" | python3 -c \
  "import sys,json; d=json.load(sys.stdin); print(d.get('tool_input',{}).get('command',''))" \
  2>/dev/null)
parse_status=$?

# Fail closed: if we could not parse the command out of a non-empty payload,
# deny rather than allow an unclassified command through.
if [[ $parse_status -ne 0 ]]; then
  [[ -z "$input" ]] && exit 0
  deny "context-mode guard could not parse the Bash command (failing closed). Re-run through ctx_execute or report the hook error."
fi

[[ -z "$cmd" ]] && exit 0

cmd_flat=$(printf '%s' "$cmd" | tr '\n' ' ')

# Allowlist first: state mutation, short OBSERVE, installs, navigation, and
# interactive commands run direct. A leading allowed verb short-circuits so a
# gather verb appearing later as an argument (e.g. a commit message) is ignored.
# Built by concatenation to keep each source line within the line-length limit;
# the assembled pattern is identical to the original single-line regex.
allow_re='^[[:space:]]*('
allow_re+='git[[:space:]]+(status|add|commit|push|pull|fetch|checkout|switch|branch|merge|rebase|reset|restore|stash|rm|mv|tag|init|clone|config|remote|worktree|rev-parse)'
allow_re+='|mkdir|mv|cp|rm|touch|chmod|chown|ln|cd|pwd|whoami|echo|printf|export|which|command[[:space:]]+-v'
allow_re+='|go[[:space:]]+(build|install|mod|fmt|generate)'
allow_re+='|npm[[:space:]]+(install|ci|run[[:space:]]+build)'
allow_re+='|make[[:space:]]+(build|install|clean))'
allow_re+='([[:space:]]|$)'
if printf '%s' "$cmd_flat" | grep -qE "$allow_re" 2>/dev/null; then
  exit 0
fi

# Denylist: context-producing / gather-and-analyze verbs the rule names.
if printf '%s' "$cmd_flat" | grep -qE \
  "(^|[[:space:]|(&;])((grep|egrep|rg|find|cat|awk|sed|nl|ls[[:space:]]+-[a-zA-Z]*R)[[:space:]]|go[[:space:]]+(test|vet)([[:space:]]|$)|git[[:space:]]+(log|diff|show|blame)([[:space:]]|$)|make[[:space:]]+(test|lint|security|coverage|mutation))" \
  2>/dev/null ||
  printf '%s' "$cmd_flat" | grep -qE "\|[[:space:]]*(head|tail|wc)([[:space:]]|$)" 2>/dev/null; then
  deny "context-mode-required.md: route this through the context-mode tool (ctx_execute / ctx_batch_execute / ctx_execute_file) so raw output stays in the sandbox. Do NOT bypass with '# ctx-ok' — the rule forbids it for context-producing commands."
fi

exit 0
