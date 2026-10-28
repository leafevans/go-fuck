package rules

import "testing"

const brewUninstallOutput = "Uninstalling /usr/local/Cellar/tbb/4.4-20160916... (118 files, 1.9M)\n" +
	"tbb 4.4-20160526, 4.4-20160722 are still installed.\n" +
	"Remove all versions with `brew uninstall --force tbb`.\n"

func TestRuleBrewUninstall(t *testing.T) {
	resetTestState(t)
	rule := Get("brew_uninstall")
	if rule == nil {
		t.Fatal("brew_uninstall rule is not registered")
	}

	assertMatch(t, rule, "brew uninstall tbb", brewUninstallOutput)
	assertMatch(t, rule, "brew rm tbb", brewUninstallOutput)
	assertMatch(t, rule, "brew remove tbb", brewUninstallOutput)

	assertNotMatch(t, rule, "brew remove gnuplot",
		"Uninstalling /usr/local/Cellar/gnuplot/5.0.4_1... (44 files, 2.3M)\n")

	assertNewCommands(t, rule, "brew uninstall tbb", brewUninstallOutput,
		[]string{"brew uninstall --force tbb"})
}
