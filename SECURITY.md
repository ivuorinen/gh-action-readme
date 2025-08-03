# Security Policy

## Supported Versions

We provide security updates for the following versions of gh-action-readme:

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| < latest| :x:                |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security issue in gh-action-readme, please report it responsibly.

### How to Report

1. **Do NOT create a public GitHub issue** for security vulnerabilities
2. Send an email to [security@ivuorinen.dev](mailto:security@ivuorinen.dev) with:
  - Description of the vulnerability
  - Steps to reproduce the issue
  - Potential impact assessment
  - Any suggested fixes (if available)

### What to Expect

- **Acknowledgment**: We'll acknowledge receipt of your report within 48 hours
- **Investigation**: We'll investigate and validate the report within 5 business days
- **Resolution**: We'll work on a fix and coordinate disclosure timeline
- **Credit**: We'll credit you in the security advisory (unless you prefer to remain anonymous)

## Security Measures

### Automated Security Scanning

We employ multiple layers of automated security scanning:

- **govulncheck**: Go-specific vulnerability scanning
- **Snyk**: Dependency vulnerability analysis
- **Trivy**: Container and filesystem security scanning
- **gitleaks**: Secrets detection and prevention
- **CodeQL**: Static code analysis
- **Dependabot**: Automated dependency updates

### Secure Development Practices

- All dependencies are regularly updated
- Security patches are prioritized
- Code is reviewed by maintainers
- CI/CD pipelines include security checks
- Container images are scanned for vulnerabilities

### Supply Chain Security

- Dependencies are pinned to specific versions
- SBOM (Software Bill of Materials) is generated for releases
- Artifacts are signed using Cosign
- Docker images are built with minimal attack surface

## Security Configuration

### For Users

When using gh-action-readme in your projects:

1. **Keep Updated**: Always use the latest version
2. **Review Permissions**: Only grant necessary GitHub token permissions
3. **Validate Inputs**: Sanitize any user-provided inputs
4. **Monitor Dependencies**: Use Dependabot or similar tools

### For Contributors

When contributing to gh-action-readme:

1. **Follow Security Guidelines**: See [CONTRIBUTING.md](CONTRIBUTING.md)
2. **Run Security Scans**: Use `make security` before submitting PRs
3. **Handle Secrets Carefully**: Never commit secrets or API keys
4. **Update Dependencies**: Keep dependencies current and secure

## Known Security Considerations

### GitHub Token Usage

gh-action-readme requires GitHub API access for dependency analysis:

- Uses read-only operations when possible
- Respects rate limits to prevent abuse
- Caches results to minimize API calls
- Never stores or logs authentication tokens

### Template Processing

Template rendering includes security measures:

- Input sanitization for user-provided data
- No execution of arbitrary code
- Limited template functions to prevent injection

## Security Tools and Commands

### Local Security Testing

```bash
# Run all security scans
make security

# Individual scans
make vulncheck  # Go vulnerability check
make snyk       # Dependency analysis
make trivy      # Filesystem scanning
make gitleaks   # Secrets detection

# Security audit
make audit      # Comprehensive dependency audit
```

### CI/CD Security

Our GitHub Actions workflows automatically run:

- Security scans on every PR and push
- Weekly scheduled vulnerability checks
- Dependency reviews for pull requests
- Container image security scanning

## Security Best Practices for Users

### GitHub Actions Usage

```yaml
# Recommended secure usage
- name: Generate README
  uses: ivuorinen/gh-action-readme@v1
  with:
    github-token: ${{ secrets.GITHUB_TOKEN }}
    # Limit token permissions in workflow
  permissions:
    contents: read
    metadata: read
```

### Local Development

```bash
# Install security tools
go install golang.org/x/vuln/cmd/govulncheck@latest
npm install -g snyk
# Install trivy: https://aquasecurity.github.io/trivy/
# Install gitleaks: https://github.com/gitleaks/gitleaks

# Run before committing
make security
```

## Incident Response

In case of a security incident:

1. **Immediate Response**: Assess and contain the issue
2. **Communication**: Notify affected users through security advisories
3. **Remediation**: Release patches and updated documentation
4. **Post-Incident**: Review and improve security measures

## Security Contact

For security-related questions or concerns:

- **Email**: [security@ivuorinen.dev](mailto:security@ivuorinen.dev)
- **PGP Key**: Available upon request
- **Response Time**: Within 48 hours for security issues

---

*This security policy is reviewed quarterly and updated as needed to reflect current best practices and threat landscape.*
