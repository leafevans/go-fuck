// Package scripts tests the release updaters and the shipped shell shims
// without network access or user state.
package scripts

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// shippedScript extensions that are packaged in release archives or executed
// by users directly. The shell shims are copied into every release archive by
// .goreleaser.yml.
var shippedScriptExtensions = map[string]bool{
	".bat":  true,
	".ps1":  true,
	".sh":   true,
	".fish": true,
}

// moduleReference matches a module path or install URL of this repository and
// captures an optional version suffix, for example "/v4" or "/cmd".
var moduleReference = regexp.MustCompile(`github\.com/leafevans/go-fuck(/v[0-9]+|/cmd)?`)

// currentModulePath returns the module path declared in go.mod, for example
// "github.com/leafevans/go-fuck/v4".
func currentModulePath(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	t.Fatal("go.mod does not declare a module path")
	return ""
}

// majorSuffix returns the "/vN" suffix of modulePath, or an empty string for
// modules without a major version suffix.
func majorSuffix(modulePath string) string {
	re := regexp.MustCompile(`/v[0-9]+$`)
	return re.FindString(modulePath)
}

// TestShippedScriptsReferenceCurrentModulePath guards the install hints in the
// shipped shell shims and updater scripts against a stale major version. The
// shims are packaged in every release archive, so a stale hint is user facing.
func TestShippedScriptsReferenceCurrentModulePath(t *testing.T) {
	modulePath := currentModulePath(t)
	major := majorSuffix(modulePath)

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read scripts directory: %v", err)
	}
	scanned := 0
	for _, entry := range entries {
		if entry.IsDir() || !shippedScriptExtensions[filepath.Ext(entry.Name())] {
			continue
		}
		scanned++
		path := entry.Name()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for lineNumber, line := range strings.Split(string(data), "\n") {
			for _, match := range moduleReference.FindAllStringSubmatch(line, -1) {
				reference := match[0]
				suffix := match[1]
				switch {
				case suffix == "/cmd":
					t.Errorf("%s:%d: %q uses the pre-v2 module path; want %q",
						path, lineNumber+1, reference, modulePath+"/cmd")
				case suffix != "" && suffix != major:
					t.Errorf("%s:%d: %q uses module suffix %q; want %q",
						path, lineNumber+1, reference, suffix, modulePath)
				}
				// An empty suffix is a repository URL such as
				// https://github.com/leafevans/go-fuck/releases, not a
				// module path, so it is valid.
			}
		}
	}
	if scanned == 0 {
		t.Fatal("no shipped scripts were scanned; the test would not guard anything")
	}
}
