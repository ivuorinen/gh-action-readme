---
name: new-theme
description: Scaffold a new documentation theme following the project's theme registration pattern from CLAUDE.md
---

The user provides a theme name (kebab-case, e.g. `dark-mode`).

Follow the "Adding a New Theme" steps from CLAUDE.md exactly:

1. Create `templates_embed/templates/themes/<theme-name>/readme.tmpl`.
    Read the default template at `templates_embed/templates/readme.tmpl` and use it as a starting point.

2. Add two constants to `appconstants/constants.go`:

    ```go
    Theme<ThemeName>        = "<theme-name>"
    TemplatePath<ThemeName> = "templates/themes/<theme-name>/readme.tmpl"
    ```

    Note: the path uses `templates/` prefix (not `templates_embed/templates/`).

3. Add a case to `internal/config.go:resolveThemeTemplate()`:

    ```go
    case appconstants.Theme<ThemeName>:
        templatePath = appconstants.TemplatePath<ThemeName>
    ```

4. Update `configThemesHandler()` in `main.go` to include the new theme name in its output.

5. Run `go build .` — report success or any compilation errors.

6. Test: `go run . gen testdata/example-action/ --theme <theme-name> --output /tmp/test-<theme-name>.md`
    Show the first 20 lines of the generated output.
