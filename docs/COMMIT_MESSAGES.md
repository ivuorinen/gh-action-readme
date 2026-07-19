# Semantic Commit Messages

This project follows [Conventional Commits](https://www.conventionalcommits.org/) specification for commit messages.

## Format

```text
<type>(<scope>): <subject>

<body>

<footer>
```

### Type

Must be one of the following:

- **feat**: A new feature
- **fix**: A bug fix
- **docs**: Documentation only changes
- **style**: Changes that do not affect the meaning of the code (white-space, formatting, etc)
- **refactor**: A code change that neither fixes a bug nor adds a feature
- **perf**: A code change that improves performance
- **test**: Adding missing tests or correcting existing tests
- **chore**: Changes to the build process or auxiliary tools
- **ci**: Changes to CI configuration files and scripts
- **build**: Changes that affect the build system or external dependencies
- **revert**: Reverts a previous commit

### Scope

The scope is optional and can be anything specifying the place of the commit change.

### Subject

The subject contains a succinct description of the change:

- Use the imperative, present tense: "change" not "changed" nor "changes"
- Don't capitalize the first letter
- No dot (.) at the end

### Examples

```text
feat: add support for AsciiDoc output format
fix: correct template rendering for empty descriptions
docs: update installation instructions
chore: prepare release v1.2.3
ci: update cosign version to v2.4.0
```

### Breaking changes

Any commit that changes an existing user-facing contract **must** be marked as
breaking, even under a `feat`/`fix`/`refactor` type. This drives the major
version bump and warns downstream users. A contract change includes:

- changing or removing a CLI command, subcommand, or flag (or altering the
  meaning of an existing positional argument);
- changing a generated output format (Markdown/HTML/JSON/AsciiDoc structure,
  JSON schema keys) or a config-file key, default, or precedence;
- changing an exported Go API signature or behavior.

Mark it in one of two ways:

- append `!` after the type/scope — `feat(validate)!: require an existing path`;
- or add a `BREAKING CHANGE:` footer describing the change and migration.

```text
feat(deps)!: remove the unimplemented `deps graph` command

BREAKING CHANGE: `deps graph` is gone; use `deps list` instead.
```

Rule of thumb: if a previously-working invocation, output, or import would break
after the change, it is a breaking change — mark it, and update the affected
docs in the same commit.

## Validation

Commit messages are validated using commitlint:

- **Pre-commit hook**: Validates commit messages before they are created (if pre-commit is installed)
- **CI/CD**: GitHub Actions workflow validates all commits in pull requests
- **Release script**: Warns if recent commits don't follow the format

## Setup

To enable local commit message validation:

```bash
# Install pre-commit hooks
make pre-commit-install

# Or manually
npm install
```

## Resources

- [Conventional Commits](https://www.conventionalcommits.org/)
- [Commitlint](https://commitlint.js.org/)
