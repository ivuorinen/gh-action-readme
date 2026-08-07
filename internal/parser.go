package internal

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

// ActionYML models the action.yml metadata (fields are updateable as schema evolves).
type ActionYML struct {
	Name        string                  `yaml:"name"`
	Description string                  `yaml:"description"`
	Inputs      map[string]ActionInput  `yaml:"inputs"`
	Outputs     map[string]ActionOutput `yaml:"outputs"`
	Runs        ActionRuns              `yaml:"runs"`
	Branding    *Branding               `yaml:"branding,omitempty"`
	Permissions PermissionMap           `yaml:"permissions,omitempty"`
	// License is the action's license identifier (e.g. "MIT", "Apache-2.0"). It is
	// not part of GitHub's action.yml schema; it is accepted as a top-level key and
	// as a `# license:` header comment, and can be detected from a repository
	// LICENSE file. Empty means unknown — templates then render no license section
	// rather than asserting one.
	License string `yaml:"license,omitempty"`
	// Add more fields as the schema evolves
}

// ActionRuns models the `runs:` section, covering every valid GitHub Actions key
// across composite, docker, and javascript action types. Steps stay as loosely
// typed maps (deep-typing steps is out of scope). Parsing is lossless for valid
// action.yml files.
type ActionRuns struct {
	Using          string            `json:"using,omitempty"           yaml:"using,omitempty"`
	Main           string            `json:"main,omitempty"            yaml:"main,omitempty"`
	Pre            string            `json:"pre,omitempty"             yaml:"pre,omitempty"`
	Post           string            `json:"post,omitempty"            yaml:"post,omitempty"`
	PreIf          string            `json:"pre-if,omitempty"          yaml:"pre-if,omitempty"`
	PostIf         string            `json:"post-if,omitempty"         yaml:"post-if,omitempty"`
	Image          string            `json:"image,omitempty"           yaml:"image,omitempty"`
	Entrypoint     string            `json:"entrypoint,omitempty"      yaml:"entrypoint,omitempty"`
	PreEntrypoint  string            `json:"pre-entrypoint,omitempty"  yaml:"pre-entrypoint,omitempty"`
	PostEntrypoint string            `json:"post-entrypoint,omitempty" yaml:"post-entrypoint,omitempty"`
	Args           []string          `json:"args,omitempty"            yaml:"args,omitempty"`
	Env            map[string]string `json:"env,omitempty"             yaml:"env,omitempty"`
	Steps          []map[string]any  `json:"steps,omitempty"           yaml:"steps,omitempty"`
}

// IsEmpty reports whether the runs section carries no entry point. Every valid
// action sets `using`; the entry-point fields cover the rare malformed cases where
// it is missing, so this stands in for the old `len(Runs) == 0` map check.
func (r ActionRuns) IsEmpty() bool {
	return r.Using == "" && len(r.Steps) == 0 && r.Image == "" && r.Main == ""
}

// PermissionMap is a scope→level permission mapping that also accepts GitHub's
// scalar shorthand (`permissions: read-all` / `write-all` / `none`). Its
// underlying type is map[string]string, so it ranges in templates and marshals to
// a JSON object exactly like a plain map — the only difference is parsing.
type PermissionMap map[string]string

// UnmarshalYAML decodes either the mapping form (`permissions:\n  contents: read`)
// or the scalar shorthand into a PermissionMap, instead of failing the whole file
// on the scalar form (goccy returns "string was used where mapping is expected").
// The scalars map to an `all` scope so they render sensibly (`all: read`).
func (p *PermissionMap) UnmarshalYAML(unmarshal func(any) error) error {
	var raw any
	if err := unmarshal(&raw); err != nil {
		return err
	}

	switch v := raw.(type) {
	case nil:
		*p = nil
	case string:
		return p.fromScalar(v)
	case map[string]any:
		m := make(PermissionMap, len(v))
		for key, val := range v {
			m[key] = fmt.Sprintf("%v", val)
		}
		*p = m
	default:
		return fmt.Errorf("invalid type %T for permissions", raw)
	}

	return nil
}

// fromScalar handles the `read-all` / `write-all` / `none` shorthand.
func (p *PermissionMap) fromScalar(v string) error {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "read-all":
		*p = PermissionMap{appconstants.PermissionScopeAll: appconstants.PermissionRead}
	case "write-all":
		*p = PermissionMap{appconstants.PermissionScopeAll: appconstants.PermissionWrite}
	case appconstants.PermissionNone, "":
		*p = PermissionMap{}
	default:
		return fmt.Errorf("invalid permissions scalar %q (want read-all, write-all, or none)", v)
	}

	return nil
}

// ActionInput represents an input parameter for a GitHub Action.
type ActionInput struct {
	Description string   `yaml:"description"`
	Required    FlexBool `yaml:"required"`
	Default     any      `yaml:"default"`
}

// FlexBool is a bool that also accepts the quoted/alternate YAML boolean forms
// GitHub tolerates (`required: "true"`, `yes`, `on`, `1` and their negatives).
// Its underlying type is bool, so it renders as true/false in templates and
// marshals as a JSON boolean — transparent everywhere except parsing.
type FlexBool bool

// UnmarshalYAML decodes a scalar into a FlexBool, accepting native booleans and
// the common string spellings rather than failing the whole file on a quoted
// value (a strict bool field returns "cannot unmarshal string into bool").
func (b *FlexBool) UnmarshalYAML(unmarshal func(any) error) error {
	var raw any
	if err := unmarshal(&raw); err != nil {
		return err
	}

	switch v := raw.(type) {
	case bool:
		*b = FlexBool(v)
	case nil:
		*b = false
	case string:
		return b.fromString(v)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		// Unquoted YAML numbers (required: 1 / required: 0) decode as numeric
		// scalars, not strings; normalize to the string form so they accept the same
		// 1/0 spellings as quoted values instead of failing the whole file.
		return b.fromString(fmt.Sprintf("%v", v))
	default:
		return fmt.Errorf("invalid type %T for required", raw)
	}

	return nil
}

// fromString sets the FlexBool from one of the accepted string spellings, returning
// an error for anything else.
func (b *FlexBool) fromString(v string) error {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "true", "yes", "on", "1":
		*b = true
	case "false", "no", "off", "0", "":
		*b = false
	default:
		return fmt.Errorf("invalid boolean value for required: %q", v)
	}

	return nil
}

// ActionOutput represents an output parameter for a GitHub Action.
type ActionOutput struct {
	Description string `yaml:"description"`
}

// Branding represents the branding configuration for a GitHub Action.
type Branding struct {
	Icon  string `yaml:"icon"`
	Color string `yaml:"color"`
}

// ParseActionYML reads and parses action.yml from given path. Warnings from the
// header-comment permission scan are discarded; use ParseActionYMLWithWarnings to
// surface them.
func ParseActionYML(path string) (*ActionYML, error) {
	action, _, err := ParseActionYMLWithWarnings(path)

	return action, err
}

// ParseActionYMLWithWarnings reads and parses action.yml, additionally returning
// non-fatal warnings. A failure to scan header-comment permissions (an unreadable
// file, or a comment line past bufio.Scanner's 64 KiB token limit) must not fail the
// whole parse, but it silently drops every comment-declared permission — so the
// caller gets a warning it can log instead of the section vanishing unannounced.
func ParseActionYMLWithWarnings(path string) (*ActionYML, []string, error) {
	var warnings []string

	// Scan the header comment block FIRST (permissions + license conventions).
	header, err := scanHeaderComments(path)
	if err != nil {
		warnings = append(warnings,
			fmt.Sprintf("could not parse header comments in %s: %v", path, err))
		header = headerComments{}
	}
	commentPermissions := header.Permissions

	// Standard YAML parsing
	f, err := os.Open(path) // #nosec G304 -- path from function parameter
	if err != nil {
		return nil, warnings, err
	}
	defer func() {
		_ = f.Close() // Ignore close error in defer
	}()
	var a ActionYML
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&a); err != nil {
		return nil, warnings, err
	}

	// Merge permissions: YAML permissions override comment permissions
	mergePermissions(&a, commentPermissions)

	// The YAML key wins over the header comment, matching the permissions precedence.
	if a.License == "" {
		a.License = header.License
	}

	return &a, warnings, nil
}

// mergePermissions combines comment and YAML permissions.
// YAML permissions take precedence when both exist.
func mergePermissions(action *ActionYML, commentPerms map[string]string) {
	if len(commentPerms) == 0 {
		return
	}
	if action.Permissions == nil {
		action.Permissions = PermissionMap(commentPerms)

		return
	}
	// A scalar shorthand or an explicit empty map is a GLOBAL, authoritative
	// statement, so comment-declared per-scope permissions must not leak past it:
	// `permissions: none` / `{}` decode to a non-nil empty map, and
	// `permissions: read-all` / `write-all` decode to a single {all: ...} entry.
	// Only a per-scope YAML map (no reserved `all` key) merges with comments,
	// keeping the documented "YAML wins per key, all unique keys kept" behavior.
	if len(action.Permissions) == 0 {
		return
	}
	if _, isGlobal := action.Permissions[appconstants.PermissionScopeAll]; isGlobal {
		return
	}
	for key, value := range commentPerms {
		if _, exists := action.Permissions[key]; !exists {
			action.Permissions[key] = value
		}
	}
}

// headerComments holds the metadata recovered from an action file's leading comment
// block. GitHub's action.yml schema has no place for these, so they are expressed as
// a comment convention and scanned here.
type headerComments struct {
	Permissions map[string]string
	License     string
}

// scanHeaderComments reads the leading comment block of an action file once and
// extracts every supported convention from it.
//
// A single-line `license:` declaration:
//
//	# license: Apache-2.0
//
// and a `permissions:` block, in either the dash or the plain mapping form, with
// optional trailing inline comments:
//
//	# permissions:
//	#   - contents: read  # Required for checking out repository
//	#   contents: read    # Alternative format without dash
//
// Scanning stops at the first non-comment line, so only the file header is read.
func scanHeaderComments(path string) (headerComments, error) {
	result := headerComments{Permissions: make(map[string]string)}

	file, err := os.Open(path) // #nosec G304 -- path from function parameter
	if err != nil {
		return result, err
	}
	defer func() {
		_ = file.Close() // Ignore close error in defer
	}()

	scanner := bufio.NewScanner(file)
	state := headerScanState{expectedItemIndent: -1}

	for scanner.Scan() {
		line := scanner.Text()
		// A UTF-8 BOM on the first line is not Unicode whitespace, so TrimSpace
		// leaves it glued to the leading "#" and the scan would abort on line 1,
		// silently dropping comment-declared permissions. Strip it.
		line = strings.TrimPrefix(line, "\ufeff")
		trimmed := strings.TrimSpace(line)

		// Stop parsing at first non-comment line
		if !strings.HasPrefix(trimmed, "#") {
			break
		}

		// Remove leading # and spaces
		content := strings.TrimPrefix(trimmed, "#")
		content = strings.TrimSpace(content)

		state.consume(line, content, &result)
	}

	if err := scanner.Err(); err != nil {
		return result, err
	}

	return result, nil
}

// headerScanState carries the cross-line state of the header comment scan: whether
// we are inside a `permissions:` block and the indent its items established.
type headerScanState struct {
	inPermissionsBlock bool
	expectedItemIndent int
}

// consume folds one header comment line into result. line is the raw line (needed for
// indent measurement) and content is that line with the leading "#" and surrounding
// whitespace removed.
func (s *headerScanState) consume(line, content string, result *headerComments) {
	// A dedent ends the permissions block, and the line that caused it belongs to
	// whatever follows — so it is re-offered to this function rather than consumed
	// by the block that just closed. Without the retry, `# license:` written after
	// an indented permission entry was silently dropped.
	for range 2 {
		if !s.consumeOnce(line, content, result) {
			return
		}
	}
}

// consumeOnce folds one line into result and reports whether the line closed a
// permissions block without being consumed, meaning the caller should offer it again
// now that the block state is cleared.
func (s *headerScanState) consumeOnce(line, content string, result *headerComments) bool {
	// Check for permissions block start.
	if content == "permissions:" {
		s.inPermissionsBlock = true
		// Items are indented relative to the content; the indent is learned from the
		// first item, so reset it here.
		s.expectedItemIndent = -1

		return false
	}

	// A `license:` line is only recognized outside a permissions block, so an
	// (invalid) "license" permission scope cannot be mistaken for a declaration.
	// First declaration wins.
	if !s.inPermissionsBlock && result.License == "" {
		if v, ok := parseHeaderScalar(content, appconstants.HeaderFieldLicense); ok {
			result.License = v

			return false
		}
	}

	// Parse permission entries.
	if s.inPermissionsBlock && content != "" {
		if processPermissionEntry(line, content, &s.expectedItemIndent, result.Permissions) {
			// A dedent ends the current permissions block but must NOT abort the whole
			// header scan — a later "# permissions:" block or a `# license:` line would
			// otherwise be silently dropped. The line was not consumed as a permission,
			// so ask the caller to re-offer it now that the block is closed.
			s.inPermissionsBlock = false
			s.expectedItemIndent = -1

			return true
		}
	}

	return false
}

// parseHeaderScalar matches a `<field>: <value>` comment line, case-insensitively on
// the field name, strips any trailing inline comment, and returns the trimmed value.
// An empty value reports false so a bare `# license:` is treated as absent.
func parseHeaderScalar(content, field string) (string, bool) {
	prefix := field + ":"
	if len(content) < len(prefix) || !strings.EqualFold(content[:len(prefix)], prefix) {
		return "", false
	}

	value := content[len(prefix):]
	if idx := strings.Index(value, "#"); idx >= 0 {
		value = value[:idx]
	}

	// Tolerate a quoted value (`# license: "MIT"`).
	value = unquoteYAMLScalarValue(strings.TrimSpace(value))
	if value == "" {
		return "", false
	}

	return value, true
}

// unquoteYAMLScalarValue strips one matching pair of surrounding single or double
// quotes from a scalar; otherwise returns the input unchanged.
func unquoteYAMLScalarValue(s string) string {
	if len(s) >= 2 {
		first, last := s[0], s[len(s)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			return s[1 : len(s)-1]
		}
	}

	return s
}

// parsePermissionLine extracts key-value from a permission line.
// Supports formats:
//   - "- contents: read  # comment"
//   - "contents: read  # comment"
func parsePermissionLine(content string) (key, value string, ok bool) {
	// Remove leading dash if present
	content = strings.TrimPrefix(content, "-")
	content = strings.TrimSpace(content)

	// Remove inline comments
	if idx := strings.Index(content, "#"); idx > 0 {
		content = strings.TrimSpace(content[:idx])
	}

	// Parse key: value
	parts := strings.SplitN(content, ":", 2)
	if len(parts) == 2 {
		key = strings.TrimSpace(parts[0])
		value = strings.TrimSpace(parts[1])
		if key != "" && value != "" {
			return key, value, true
		}
	}

	return "", "", false
}

// isValidPermissionValue reports whether v is a GitHub Actions permission level
// (read, write, or none); the match is case-insensitive.
func isValidPermissionValue(v string) bool {
	switch strings.ToLower(v) {
	case appconstants.PermissionRead, appconstants.PermissionWrite, appconstants.PermissionNone:
		return true
	default:
		return false
	}
}

// processPermissionEntry processes a single line in the permissions block.
// Returns true if the line is dedented out of the current block (the caller
// should end the block but keep scanning), false otherwise.
func processPermissionEntry(line, content string, expectedItemIndent *int, permissions map[string]string) bool {
	// Get the indent of the content (after removing #). Strip both spaces and
	// tabs so a tab-indented comment item is recognized as indentation rather
	// than content (which would otherwise truncate the permissions block early).
	lineAfterHash := strings.TrimPrefix(line, "#")
	contentIndent := len(lineAfterHash) - len(strings.TrimLeft(lineAfterHash, " \t"))

	// Set expected indent on first item
	if *expectedItemIndent == -1 {
		*expectedItemIndent = contentIndent
	}

	// If dedented relative to expected item indent, we've left the permissions block
	if contentIndent < *expectedItemIndent {
		return true
	}

	// Parse permission line and add to map only when the value is a real
	// permission level. This rejects a prose line with a colon inside the
	// permissions comment block (e.g. "Note: needs network access") that would
	// otherwise be recorded as a bogus permission.
	if key, value, ok := parsePermissionLine(content); ok && isValidPermissionValue(value) {
		permissions[key] = value
	}

	return false
}

// shouldIgnoreDirectory checks if a directory name exactly matches the ignore list.
func shouldIgnoreDirectory(dirName string, ignoredDirs []string) bool {
	for _, ignored := range ignoredDirs {
		if dirName == ignored {
			return true
		}
	}

	return false
}

// actionFileWalker encapsulates the logic for walking directories and finding action files.
type actionFileWalker struct {
	ignoredDirs []string
	actionFiles []string
}

// walkFunc is the callback function for filepath.Walk.
func (w *actionFileWalker) walkFunc(path string, info os.FileInfo, err error) error {
	if err != nil {
		if os.IsPermission(err) {
			if info != nil && info.IsDir() {
				return filepath.SkipDir
			}

			return nil
		}

		return err
	}

	if info.IsDir() {
		if shouldIgnoreDirectory(info.Name(), w.ignoredDirs) {
			return filepath.SkipDir
		}

		return nil
	}

	// Check for action.yml or action.yaml files
	if isActionFileName(info.Name()) {
		w.actionFiles = append(w.actionFiles, path)
	}

	return nil
}

// isActionFileName reports whether name is a GitHub Actions metadata file name.
// The match is case-SENSITIVE on purpose: GitHub only loads lowercase action.yml /
// action.yaml, so treating "Action.YML" as an action would generate documentation —
// and a uses: statement — for a file GitHub will never resolve. Both discovery paths
// share this predicate so recursive and non-recursive runs agree on the action set.
func isActionFileName(name string) bool {
	return name == appconstants.ActionFileNameYML || name == appconstants.ActionFileNameYAML
}

// DiscoverActionFiles finds action.yml and action.yaml files in the given directory.
// This consolidates the file discovery logic from both generator.go and dependencies/parser.go.
func DiscoverActionFiles(dir string, recursive bool, ignoredDirs []string) ([]string, error) {
	// Check if dir exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", dir)
	}

	if recursive {
		walker := &actionFileWalker{ignoredDirs: ignoredDirs}
		if err := filepath.Walk(dir, walker.walkFunc); err != nil {
			return nil, fmt.Errorf("failed to walk directory %s: %w", dir, err)
		}

		return walker.actionFiles, nil
	}

	// Check only the specified directory (non-recursive)
	return DiscoverActionFilesNonRecursive(dir), nil
}

// DiscoverActionFilesNonRecursive finds action files (action.yml or action.yaml) in a single directory.
// This is exported for use by other packages that need to discover action files.
func DiscoverActionFilesNonRecursive(dir string) []string {
	var actionFiles []string
	for _, filename := range []string{appconstants.ActionFileNameYML, appconstants.ActionFileNameYAML} {
		path := filepath.Join(dir, filename)
		if _, err := os.Stat(path); err == nil {
			actionFiles = append(actionFiles, path)
		}
	}

	return actionFiles
}
