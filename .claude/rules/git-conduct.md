# Git Conduct

Never pass `--no-verify` or its short form `-n` to `git commit` or `git push`.
The pre-commit and commit-msg hooks are required; fix the underlying failure
instead of skipping them.

Never force-push. This includes `--force`, `-f`, `--force-with-lease`, and the
`+refspec` shorthand (`git push origin +main`). Recover with a new branch or a
revert commit.

Never override the hooks path. Do not run `git -c core.hooksPath=...` to point
git at an alternate (or empty) hooks directory.

These prohibitions are enforced by `.claude/hooks/pre-bash-dangerous-guard.sh`.
