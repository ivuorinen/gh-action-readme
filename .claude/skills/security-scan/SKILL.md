---
name: security-scan
description: Run all project security scans (make security) and summarize findings by tool and severity
---

Run `make security` from the project root. This executes:

- `govulncheck ./...` — Go vulnerability database check
- `gosec ./...` — Go static security analysis
- `trivy fs . --severity HIGH,CRITICAL` — filesystem vulnerability scan
- `gitleaks detect --source . --verbose` — secrets detection

For each tool, summarize:

- Count of findings by severity (CRITICAL, HIGH, MEDIUM, LOW)
- Full details for any CRITICAL or HIGH findings: file:line, rule ID, description, remediation

Final verdict:

- **CLEAN**: no HIGH or CRITICAL findings across all tools
- **NEEDS ATTENTION**: list specific items requiring action, ordered by severity
