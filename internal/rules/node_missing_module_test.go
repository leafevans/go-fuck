package rules

import "testing"

// nodeMissingModuleOutput is the output of Node.js for a missing package.
func nodeMissingModuleOutput() string {
	return "node:internal/modules/cjs/loader:1479\n" +
		"  throw err;\n" +
		"  ^\n" +
		"\n" +
		"Error: Cannot find module 'expresss'\n" +
		"Require stack:\n" +
		"- C:\\repo\\app.js\n" +
		"    at Module._resolveFilename (node:internal/modules/cjs/loader:1476:15)\n" +
		"  code: 'MODULE_NOT_FOUND'\n"
}

// nodeMissingFileOutput is the output of Node.js for a missing relative file.
func nodeMissingFileOutput() string {
	return "Error: Cannot find module './missing.js'\n" +
		"Require stack:\n" +
		"- C:\\repo\\app.js\n"
}

func TestRuleNodeMissingModule(t *testing.T) {
	resetTestState(t)
	rule := Get("node_missing_module")
	if rule == nil {
		t.Fatal("node_missing_module rule is not registered")
	}

	assertMatch(t, rule, "node app.js", nodeMissingModuleOutput())
	assertNotMatch(t, rule, "deno app.js", nodeMissingModuleOutput())
	assertNotMatch(t, rule, "node app.js", "SyntaxError: Unexpected token")

	assertNewCommands(t, rule, "node app.js", nodeMissingModuleOutput(),
		[]string{"npm install expresss && node app.js"})
	assertNewCommands(t, rule, "node app.js", nodeMissingFileOutput(), nil)
}
