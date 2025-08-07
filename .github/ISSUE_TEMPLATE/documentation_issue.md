---
name: Documentation quality issue
about: Report problems with generated documentation output
title: '[DOCS] '
labels: documentation, quality
assignees: ivuorinen

---

**Describe the documentation issue**
A clear description of what's wrong with the generated documentation.

## Theme and format used

- Theme: [github, gitlab, minimal, professional, default]
- Output format: [md, html, json, asciidoc]
- Command used: `gh-action-readme gen [your flags here]`

**Sample action.yml**
The action.yml input that produces the problematic documentation:

```yaml
# Paste your action.yml content here
```

**Current generated output**
The problematic documentation that was generated:

```markdown
# Paste the current output here (or relevant excerpt)
```

**Expected documentation**
What the documentation should look like instead:

```markdown
# Paste what you expected here
```

**Issue category**
What type of documentation problem is this?

- [ ] Missing information (inputs, outputs, etc.)
- [ ] Incorrect formatting/rendering
- [ ] Broken links or references
- [ ] Poor table layout
- [ ] Badge/shield issues
- [ ] Code block formatting
- [ ] Template logic error
- [ ] Theme-specific styling problem
- [ ] Other: ___________

## Environment information

- OS: [e.g. macOS 14.1, Ubuntu 22.04, Windows 11]
- gh-action-readme version: [run `gh-action-readme version`]

**Additional context**
Add any other context about the documentation quality issue. Include:

- Screenshots if it's a visual formatting problem
- Links to similar actions with good documentation
- Specific markdown/HTML sections that are problematic
