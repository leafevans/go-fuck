package consts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestModulePathMatchesMajorVersion guards that the Go module path carries the
// major version of Version. Go resolves "go install module@latest" against the
// module path, so a version series whose module path lacks the /vN suffix never
// resolves: the command silently installs the newest release of the older
// series instead.
func TestModulePathMatchesMajorVersion(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	module := ""
	for line := range strings.SplitSeq(string(data), "\n") {
		if rest, ok := strings.CutPrefix(line, "module "); ok {
			module = strings.TrimSpace(rest)
			break
		}
	}
	if module == "" {
		t.Fatal("go.mod does not declare a module path")
	}
	major, _, _ := strings.Cut(Version, ".")
	suffix := ""
	if major != "0" && major != "1" {
		suffix = "/v" + major
	}
	if !strings.HasSuffix(module, suffix) {
		t.Fatalf("module path %q does not end in %q for version %q", module, suffix, Version)
	}
}
