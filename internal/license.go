package internal

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

// reLicenseFileName matches the conventional repository license file names,
// case-insensitively: LICENSE / LICENSE / COPYING, bare or with a .md/.txt/.rst
// suffix. A pattern is used rather than a literal name list so both the American and
// British spellings are covered by one expression that cannot drift apart.
var reLicenseFileName = regexp.MustCompile(`(?i)^(licen[cs]e|copying)(\.(md|txt|rst))?$`)

// licenseFileRank orders candidate file names so the choice is deterministic when a
// repository contains more than one match. Lower sorts first: the LICENSE/LICENSE
// family outranks COPYING, and a bare name outranks a suffixed one.
func licenseFileRank(name string) int {
	lower := strings.ToLower(name)

	rank := 0
	if strings.HasPrefix(lower, "copying") {
		rank += 10
	}
	if strings.Contains(lower, ".") {
		rank++
	}

	return rank
}

// licensePatterns maps a recognizable phrase from a licence's own text to its SPDX
// identifier. Ordered most-specific first: "GNU Affero" must be tested before "GNU
// General Public", and the Apache/BSD/MIT titles are distinctive enough to stand
// alone. Only the license file's opening lines are searched, where the title lives.
// The (?s) flag is essential: a license title and its version routinely sit on
// separate lines ("Apache License\nVersion 2.0, January 2004"), and without it `.`
// stops at the newline and the match silently fails.
var licensePatterns = []struct {
	pattern *regexp.Regexp
	spdx    string
}{
	{regexp.MustCompile(`(?is)GNU AFFERO GENERAL PUBLIC LICENSE`), appconstants.SPDXAGPL3},
	{regexp.MustCompile(`(?is)GNU LESSER GENERAL PUBLIC LICENSE`), appconstants.SPDXLGPL3},
	{regexp.MustCompile(`(?is)GNU GENERAL PUBLIC LICENSE.{0,40}Version 2`), appconstants.SPDXGPL2},
	{regexp.MustCompile(`(?is)GNU GENERAL PUBLIC LICENSE`), appconstants.SPDXGPL3},
	{regexp.MustCompile(`(?is)Apache License.{0,40}Version 2\.0`), appconstants.SPDXApache2},
	{regexp.MustCompile(`(?is)Mozilla Public License.{0,40}2\.0`), appconstants.SPDXMPL2},
	{regexp.MustCompile(`(?is)BSD 3-Clause`), appconstants.SPDXBSD3Clause},
	{regexp.MustCompile(`(?is)BSD 2-Clause`), appconstants.SPDXBSD2Clause},
	{regexp.MustCompile(`(?is)\bThe Unlicense\b`), appconstants.SPDXUnlicense},
	{regexp.MustCompile(`(?is)ISC License`), appconstants.SPDXISC},
	// MIT last: "MIT License" is a short, generic phrase that can appear inside
	// another licence's preamble, so every distinctive title is tested first.
	{regexp.MustCompile(`(?is)\bMIT License\b`), "MIT"},
}

// reSPDXIdentifier matches an explicit SPDX tag, which is authoritative when present.
var reSPDXIdentifier = regexp.MustCompile(`(?i)SPDX-License-Identifier:\s*([A-Za-z0-9.+-]+)`)

// licenseSniffLines bounds how much of a license file is read. The title and any
// SPDX tag appear in the opening lines; reading further would only add cost.
const licenseSniffLines = 20

// findLicenseFile returns the path to the repository's license file and its name,
// or empty strings when none is present. The directory is listed once and matched
// case-insensitively so "license.md" is found as readily as "LICENSE".
func findLicenseFile(repoRoot string) (path, name string) {
	if repoRoot == "" {
		return "", ""
	}

	entries, err := os.ReadDir(repoRoot)
	if err != nil {
		return "", ""
	}

	// Collect every match, then take the best-ranked one so the result does not
	// depend on filesystem ordering when several license files coexist.
	best := ""
	for _, entry := range entries {
		if entry.IsDir() || !reLicenseFileName.MatchString(entry.Name()) {
			continue
		}
		if best == "" || licenseFileRank(entry.Name()) < licenseFileRank(best) {
			best = entry.Name()
		}
	}

	if best == "" {
		return "", ""
	}

	return filepath.Join(repoRoot, best), best
}

// detectLicense identifies the license of the repository at repoRoot by reading its
// license file. An explicit SPDX-License-Identifier tag wins; otherwise the license
// title is matched. Returns "" when no file exists or the text is unrecognized —
// callers must treat that as unknown rather than substituting a default.
func detectLicense(repoRoot string) string {
	path, _ := findLicenseFile(repoRoot)
	if path == "" {
		return ""
	}

	file, err := os.Open(path) // #nosec G304 -- path built from a directory listing of repoRoot
	if err != nil {
		return ""
	}
	defer func() {
		_ = file.Close()
	}()

	var head strings.Builder
	scanner := bufio.NewScanner(file)
	for i := 0; i < licenseSniffLines && scanner.Scan(); i++ {
		line := scanner.Text()
		if match := reSPDXIdentifier.FindStringSubmatch(line); match != nil {
			return match[1]
		}
		head.WriteString(line)
		head.WriteString("\n")
	}
	if scanner.Err() != nil {
		return ""
	}

	text := head.String()
	for _, entry := range licensePatterns {
		if entry.pattern.MatchString(text) {
			return entry.spdx
		}
	}

	return ""
}

// resolveLicense applies the license precedence chain, highest priority first:
//
//  1. config `license:` — explicit operator override
//  2. action.yml `license:` key
//  3. `# license:` header comment (already folded into action.License by the parser)
//  4. detection from the repository's LICENSE file
//
// An empty result means unknown; templates render no license section rather than
// asserting one.
func resolveLicense(action *ActionYML, config *AppConfig, repoRoot string) string {
	if config != nil && config.License != "" {
		return config.License
	}
	if action != nil && action.License != "" {
		return action.License
	}

	return detectLicense(repoRoot)
}
