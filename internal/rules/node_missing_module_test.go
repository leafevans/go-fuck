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

// nodeMissingPackageESMOutput is the ESM output of Node.js for a missing
// package.
func nodeMissingPackageESMOutput() string {
	return "node:internal/modules/package_json_reader:301\n" +
		"  throw new ERR_MODULE_NOT_FOUND(packageName, fileURLToPath(base), null);\n" +
		"        ^\n" +
		"\n" +
		"Error [ERR_MODULE_NOT_FOUND]: Cannot find package 'expresss' imported from C:\\repo\\app.mjs\n" +
		"    at Object.getPackageJSONURL (node:internal/modules/package_json_reader:301:9)\n"
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
	assertNewCommands(t, rule, "node app.mjs", nodeMissingPackageESMOutput(),
		[]string{"npm install expresss && node app.mjs"})
	assertNewCommands(t, rule, "node app.js", nodeMissingFileOutput(), nil)
}
