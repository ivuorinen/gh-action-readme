# README Protection

Never overwrite or modify the repository-root `README.md`, whether through an
Edit/Write tool call or a Bash command (redirection, `cp`, `mv`, `tee`,
`sed -i`, `git checkout`, `git restore`). The name match is case-insensitive.

For testing generation output, always write to `/tmp/` or `testdata/`.

Enforced by `.claude/hooks/pre-readme-guard.sh` (Edit/Write) and
`.claude/hooks/pre-bash-dangerous-guard.sh` (Bash).
