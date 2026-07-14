---
paths:
  - "**/*.go"
---

# GitHub Token Accessor

Always read the GitHub token via `internal.GetGitHubToken(config)`.
Never read `config.GitHubToken` directly, except inside the config
loader, merge, and redaction sites that legitimately set or scrub the field.
