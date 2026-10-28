package rules

import "testing"

// goUnknownFlagOutput is the output of the Go toolchain for an unknown flag.
func goUnknownFlagOutput() string {
	return "flag provided but not defined: -raec\n" +
		"usage: go build [-o output] [build flags] [packages]\n" +
		"Run 'go help build' for details.\n"
}

// goUnknownFlagRelaseOutput is the output for a typo without a similar flag.
func goUnknownFlagRelaseOutput() string {
	return "flag provided but not defined: -relase\n" +
		"usage: go build [-o output] [build flags] [packages]\n" +
		"Run 'go help build' for details.\n"
}

// goBuildHelpOutput is an excerpt of "go help build" with the build flags.
func goBuildHelpOutput() string {
	return "Build compiles the packages named by the import paths.\n" +
		"\n" +
		"\t-a\n" +
		"\t\tforces rebuilding of packages\n" +
		"\t-race\n" +
		"\t\tenable race detector\n" +
		"\t-o file\n" +
		"\t\tforce build to write the resulting executable\n"
}

// withGoHelpOutput stubs the Go help output for one test.
func withGoHelpOutput(t *testing.T, help string) {
	t.Helper()
	original := goHelpOutput
	goHelpOutput = func(string) string { return help }
	t.Cleanup(func() { goHelpOutput = original })
}

func TestRuleGoUnknownFlag(t *testing.T) {
	resetTestState(t)
	withGoHelpOutput(t, goBuildHelpOutput())
	rule := Get("go_unknown_flag")
	if rule == nil {
		t.Fatal("go_unknown_flag rule is not registered")
	}

	assertMatch(t, rule, "go build --raec", goUnknownFlagOutput())
	assertNotMatch(t, rule, "cargo build --raec", goUnknownFlagOutput())
	assertNotMatch(t, rule, "go build -race", "ok  demo  0.001s")

	assertNewCommands(t, rule, "go build --raec", goUnknownFlagOutput(), []string{"go build -race"})
	assertNewCommands(t, rule, "go build --relase", goUnknownFlagRelaseOutput(), nil)
}
