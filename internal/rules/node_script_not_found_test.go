package rules

import (
	"os"
	"path/filepath"
	"testing"
)

// nodeCannotFindModuleOutput is the Node.js error for a missing script file.
const nodeCannotFindModuleOutput = "Error: Cannot find module '/proj/mian.js'\n" +
	"    at Module._resolveFilename (node:internal/modules/cjs/loader:1476:15)\n" +
	"  code: 'MODULE_NOT_FOUND'\n"

// stubNodeDirectoryFiles replaces the file list of the rule.
func stubNodeDirectoryFiles(t *testing.T, names []string) {
	t.Helper()
	original := nodeDirectoryFiles
	nodeDirectoryFiles = func(string) []string { return names }
	t.Cleanup(func() { nodeDirectoryFiles = original })
}

func TestRuleNodeScriptNotFound(t *testing.T) {
	resetTestState(t)
	rule := Get("node_script_not_found")
	if rule == nil {
		t.Fatal("node_script_not_found rule is not registered")
	}

	stubNodeDirectoryFiles(t, []string{"main.js", "lib.js"})
	assertMatch(t, rule, "node mian.js", nodeCannotFindModuleOutput)
	assertNotMatch(t, rule, "node mian.js", "")
	assertNotMatch(t, rule, "node mian.js", "Error: Cannot find module '/proj/mian.js'\n")
	assertNotMatch(t, rule, "python mian.py", nodeCannotFindModuleOutput)
	assertNewCommands(t, rule, "node mian.js", nodeCannotFindModuleOutput,
		[]string{"node main.js"})

	stubNodeDirectoryFiles(t, []string{"main.js"})
	assertNotMatch(t, rule, "node completely-different.js",
		"Error: Cannot find module '/proj/completely-different.js'\n  code: 'MODULE_NOT_FOUND'\n")
}

func TestNodePathHelpers(t *testing.T) {
	cases := []struct{ path, base, dir string }{
		{"/proj/main.js", "main.js", "/proj"},
		{"C:/proj/main.js", "main.js", "C:/proj"},
		{"main.js", "main.js", ""},
	}
	for _, testCase := range cases {
		if got := nodePathBase(testCase.path); got != testCase.base {
			t.Fatalf("nodePathBase(%q) = %q", testCase.path, got)
		}
		if got := nodePathDir(testCase.path); got != testCase.dir {
			t.Fatalf("nodePathDir(%q) = %q", testCase.path, got)
		}
	}

	for _, path := range []string{"/proj/main.js", "C:/proj/main.js"} {
		if !nodeAbsolutePath(path) {
			t.Fatalf("nodeAbsolutePath(%q) = false", path)
		}
	}
	for _, path := range []string{"main.js", "C:main.js", "proj/main.js"} {
		if nodeAbsolutePath(path) {
			t.Fatalf("nodeAbsolutePath(%q) = true", path)
		}
	}

	for _, name := range []string{"main.js", "main.MJS", "main.cjs", "main.json", "main.ts"} {
		if !nodeScriptName(name) {
			t.Fatalf("nodeScriptName(%q) = false", name)
		}
	}
	for _, name := range []string{"main.py", "README", ".txt"} {
		if nodeScriptName(name) {
			t.Fatalf("nodeScriptName(%q) = true", name)
		}
	}
}

func TestNodeScriptNotFoundEdges(t *testing.T) {
	resetTestState(t)
	rule := Get("node_script_not_found")
	stubNodeDirectoryFiles(t, []string{"main.js"})

	// A relative module path is not corrected.
	assertNotMatch(t, rule, "node mian.js",
		"Error: Cannot find module 'mian.js'\n  code: 'MODULE_NOT_FOUND'\n")
	// A missing module without a script extension is not a missing script.
	assertNotMatch(t, rule, "node mian.py",
		"Error: Cannot find module '/proj/mian.py'\n  code: 'MODULE_NOT_FOUND'\n")
	// The MODULE_NOT_FOUND line without the module line is not enough.
	assertNotMatch(t, rule, "node mian.js", "code: 'MODULE_NOT_FOUND'\n")
	// No close file means no suggestion.
	stubNodeDirectoryFiles(t, []string{"unrelated.js"})
	assertNotMatch(t, rule, "node mian.js", nodeCannotFindModuleOutput)
}

func TestNodeDirectoryFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.js"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	names := nodeDirectoryFiles(dir)
	if len(names) != 1 || names[0] != "main.js" {
		t.Fatalf("nodeDirectoryFiles() = %q", names)
	}
	if names := nodeDirectoryFiles(filepath.Join(dir, "missing")); names != nil {
		t.Fatalf("nodeDirectoryFiles(missing) = %q", names)
	}
}
