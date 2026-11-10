package rules

import "testing"

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
