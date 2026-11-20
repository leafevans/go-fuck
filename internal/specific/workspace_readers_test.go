package specific

import (
	"os"
	"slices"
	"testing"
)

func TestReadWorkspaceFiles(t *testing.T) {
	for _, tc := range []struct {
		name  string
		read  func() []string
		files map[string]string
		want  []string
	}{
		{"missing makefile", readMakeTargets, nil, nil},
		{"makefile search order", readMakeTargets, map[string]string{"GNUmakefile": "first:\n", "makefile": "second:\n", "Makefile": "third:\n"}, []string{"first"}},
		{"fallback makefile", readMakeTargets, map[string]string{"makefile": "fallback:\n"}, []string{"fallback"}},
		{"missing compose", readComposeServices, nil, nil},
		{"compose search order", readComposeServices, map[string]string{"compose.yaml": "services:\n  first:\n", "compose.yml": "services:\n  second:\n"}, []string{"first"}},
		{"fallback compose", readComposeServices, map[string]string{"docker-compose.yml": "services:\n  fallback:\n"}, []string{"fallback"}},
		{"missing package", readPackageScripts, nil, nil},
		{"invalid package", readPackageScripts, map[string]string{"package.json": "{"}, nil},
		{"invalid script value", readPackageScripts, map[string]string{"package.json": `{"scripts":{"build":42}}`}, nil},
		{"missing scripts", readPackageScripts, map[string]string{"package.json": `{"name":"demo"}`}, nil},
		{"null scripts", readPackageScripts, map[string]string{"package.json": `{"scripts":null}`}, nil},
		{"sorted scripts", readPackageScripts, map[string]string{"package.json": `{"scripts":{"test":"go test","build":"go build"}}`}, []string{"build", "test"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			for name, content := range tc.files {
				if err := os.WriteFile(name, []byte(content), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if got := tc.read(); !slices.Equal(got, tc.want) {
				t.Fatalf("workspace reader = %q, want %q", got, tc.want)
			}
		})
	}
}
