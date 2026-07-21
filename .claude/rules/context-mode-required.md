# Context-Mode Tools Required

Route context-producing commands through the context-mode tool
(`ctx_execute` / `ctx_batch_execute` / `ctx_execute_file`) so raw output stays
in the sandbox and only a derived summary enters the conversation.

The `.claude/hooks/pre-bash-context-mode-guard.sh` hook blocks these verbs in
raw Bash:

- File readers and scanners: `grep`, `egrep`, `rg`, `find`, `cat`, `awk`,
  `sed`, `nl`, `jq`, `sort`, `uniq`, `less`, `more`, `bat`, `tree`, `head`,
  `tail`, `wc`, `column`, `xxd`, `hexdump`, `od`, and `ls -R`.
- Go analysis: `go test`, `go vet`.
- Git history: `git log`, `git diff`, `git show`, `git blame`.
- Make targets: `make test`, `make lint`, `make security`, `make coverage`,
  `make mutation`.
- Log dumps: `docker logs`, `kubectl logs`, and interpreter one-liners that
  read files (`python3 -c '... open(...).read() ...'`).

The hook evaluates each segment of a chained command independently, so a
leading state-mutation verb (`cd . && go test`) does not exempt a later
context-producer.

When running a `/goal` (or any goal-driven task): if the context-mode tools are
not available in the session, treat the task as **failed** and stop. Report the
unavailability as the failure reason.
