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
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "module ") {
			module = strings.TrimSpace(strings.TrimPrefix(line, "module "))
			break
		}
	}
	if module == "" {
		t.Fatal("go.mod does not declare a module path")
	}
	major := strings.SplitN(Version, ".", 2)[0]
	suffix := ""
	if major != "0" && major != "1" {
		suffix = "/v" + major
	}
	if !strings.HasSuffix(module, suffix) {
		t.Fatalf("module path %q does not end in %q for version %q", module, suffix, Version)
	}
}
