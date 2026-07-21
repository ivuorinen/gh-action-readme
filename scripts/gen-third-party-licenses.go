//go:build ignore

// Command gen-third-party-licenses folds the license files that
// `go-licenses save` wrote into a directory tree into a single
// THIRD_PARTY_LICENSES.md, annotating each with its module version from
// `go list -m all`. Run via `make licenses`.
//
// Usage: go run scripts/gen-third-party-licenses.go <save-dir> <output.md>
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type module struct {
	path    string
	version string
	file    string
	text    string
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: gen-third-party-licenses <save-dir> <output.md>")
		os.Exit(1)
	}
	if err := run(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(saveDir, outPath string) error {
	versions, err := moduleVersions()
	if err != nil {
		return err
	}
	mods, err := collectLicenses(saveDir, versions)
	if err != nil {
		return err
	}
	sort.Slice(mods, func(i, j int) bool { return mods[i].path < mods[j].path })
	if err := os.WriteFile(outPath, []byte(render(mods)), 0o644); err != nil {
		return err
	}
	fmt.Printf("wrote %s covering %d modules\n", outPath, len(mods))
	return nil
}

// moduleVersions maps each dependency's module path to its resolved version.
func moduleVersions() (map[string]string, error) {
	out, err := exec.Command("go", "list", "-m", "-f", "{{.Path}}\t{{.Version}}", "all").Output()
	if err != nil {
		return nil, fmt.Errorf("go list -m all: %w", err)
	}
	versions := map[string]string{}
	for line := range strings.SplitSeq(strings.TrimSpace(string(out)), "\n") {
		if path, version, ok := strings.Cut(line, "\t"); ok && version != "" {
			versions[path] = version
		}
	}
	return versions, nil
}

// collectLicenses walks the go-licenses save tree; each license file's parent
// directory (relative to saveDir) is the dependency's module path.
func collectLicenses(saveDir string, versions map[string]string) ([]module, error) {
	var mods []module
	err := filepath.WalkDir(saveDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(saveDir, path)
		if err != nil {
			return err
		}
		modPath := filepath.ToSlash(filepath.Dir(rel))
		text, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		mods = append(mods, module{
			path:    modPath,
			version: versions[modPath],
			file:    filepath.Base(path),
			text:    strings.TrimSpace(string(text)),
		})
		return nil
	})
	return mods, err
}

func render(mods []module) string {
	var b strings.Builder
	b.WriteString("# Third-Party Licenses\n\n")
	b.WriteString("This binary statically links the Go modules listed below. Each module's " +
		"license text is reproduced here to satisfy its redistribution terms. " +
		"Regenerate with `make licenses`.\n\n")
	fmt.Fprintf(&b, "Modules: %d\n", len(mods))
	for _, m := range mods {
		fmt.Fprintf(&b, "\n---\n\n## %s %s\n\n_License file: %s_\n\n```\n%s\n```\n",
			m.path, m.version, m.file, m.text)
	}
	return b.String()
}
