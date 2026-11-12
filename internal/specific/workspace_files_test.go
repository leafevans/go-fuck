package specific

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestWorkspaceFileVars(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"Makefile":     "build:\n\tgo build ./...\n\ntest: build\n\tgo test ./...\n",
		"compose.yaml": "services:\n  api:\n    image: example/api\n  worker:\n    image: example/worker\n",
		"package.json": `{"scripts":{"build":"tsc","lint":"eslint ."}}`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(dir)

	if got := MakeTargets(); !slices.Contains(got, "build") || !slices.Contains(got, "test") {
		t.Fatalf("MakeTargets() = %q", got)
	}
	if got := ComposeServices(); !slices.Contains(got, "api") || !slices.Contains(got, "worker") {
		t.Fatalf("ComposeServices() = %q", got)
	}
	if got := PackageScripts(); !slices.Contains(got, "build") || !slices.Contains(got, "lint") {
		t.Fatalf("PackageScripts() = %q", got)
	}
}

func TestValidMakeTarget(t *testing.T) {
	for _, name := range []string{"build", "test-all", "deploy/prod", "a_b", "v1.2"} {
		if !validMakeTarget(name) {
			t.Fatalf("validMakeTarget(%q) = false", name)
		}
	}
	for _, name := range []string{"", ".PHONY", "$(TARGET)", "bad target", "bad%name", "#comment"} {
		if validMakeTarget(name) {
			t.Fatalf("validMakeTarget(%q) = true", name)
		}
	}
}
