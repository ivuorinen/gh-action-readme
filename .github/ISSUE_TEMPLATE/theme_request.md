---
name: Theme request
about: Request a new theme or theme improvements
title: '[THEME] '
labels: enhancement, theme
assignees: ivuorinen

---

**Theme request type**
What kind of theme enhancement is this?

- [ ] New theme for a specific platform/use case
- [ ] Improvements to existing theme
- [ ] Theme bug fix
- [ ] Theme customization options

**Target platform/use case**
What platform or scenario is this theme designed for?

- [ ] GitHub Marketplace optimization
- [ ] GitLab CI/CD documentation
- [ ] Azure DevOps integration
- [ ] Internal/Enterprise documentation
- [ ] Minimal/lightweight documentation
- [ ] Mobile-friendly documentation
- [ ] Other: ___________

**Visual requirements**
Describe the desired appearance and formatting:

**Badges and shields:**

- [ ] Platform-specific badges
- [ ] Version/release shields  
- [ ] Build status badges
- [ ] License/security badges
- [ ] Custom badge styling

**Layout preferences:**

- [ ] Table of contents
- [ ] Collapsible sections
- [ ] Multi-column layout
- [ ] Sidebar navigation
- [ ] Card-based layout
- [ ] Minimal typography

**Color scheme:**

- [ ] Dark theme
- [ ] Light theme
- [ ] Platform-specific colors (GitHub green, GitLab orange, etc.)
- [ ] Corporate/brand colors
- [ ] High contrast for accessibility

**Sample action.yml files**
Provide example action.yml files to test the theme with:

```yaml
# Example 1: Simple action
name: Simple Action
description: A basic action example
inputs:
  input1:
    description: First input
    required: true
outputs:
  output1:
    description: First output
runs:
  using: node20
  main: index.js
```

```yaml
# Example 2: Complex action (if applicable)
# Add more complex action.yml if theme needs to handle advanced cases
```

**Reference examples**
Provide links or descriptions of documentation that exemplifies the desired style:

- [ ] Link to well-styled action documentation
- [ ] Screenshots of desired layout
- [ ] Reference to design system or style guide
- [ ] Examples from other tools/platforms

**Expected output preview**
Mock up what the generated documentation should look like:

```markdown
# Show expected theme output here
# Include formatting, badges, layout structure
```

**Existing theme starting point**
Which existing theme should this be based on?

- [ ] Start from scratch
- [ ] Extend `github` theme
- [ ] Extend `gitlab` theme  
- [ ] Extend `minimal` theme
- [ ] Extend `professional` theme
- [ ] Extend `default` theme

**Additional context**
Add any other context about the theme request:

- Specific markdown features needed
- Integration requirements
- Performance considerations
- Accessibility requirements
