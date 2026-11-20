package rules

import "testing"

// powerShellSetLocationOutput is the error of Set-Location for a missing
// directory.
const powerShellSetLocationOutput = "Set-Location : Cannot find path \x27/root/missing\x27 because it does not exist.\n"

// stubPowerShellPathMissing replaces the path check of the rule.
func stubPowerShellPathMissing(t *testing.T, missing bool) {
	t.Helper()
	original := powerShellPathMissing
	powerShellPathMissing = func(string) bool { return missing }
	t.Cleanup(func() { powerShellPathMissing = original })
}

func TestRulePowerShellSetLocation(t *testing.T) {
	resetTestState(t)
	rule := Get("powershell_set_location")
	if rule == nil {
		t.Fatal("powershell_set_location rule is not registered")
	}

	stubPowerShellPathMissing(t, true)
	assertMatch(t, rule, "Set-Location \x27/root/missing\x27", powerShellSetLocationOutput)
	assertMatch(t, rule, "sl \x27/root/missing\x27", powerShellSetLocationOutput)
	assertMatch(t, rule, "chdir \x27/root/missing\x27", powerShellSetLocationOutput)
	assertMatch(t, rule, "Push-Location \x27/root/missing\x27", powerShellSetLocationOutput)
	assertMatch(t, rule, "pushd \x27/root/missing\x27", powerShellSetLocationOutput)
	assertNotMatch(t, rule, "cd \x27/root/missing\x27", powerShellSetLocationOutput)
	assertNotMatch(t, rule, "Set-Location \x27/root/missing\x27", "")
	for _, path := range []string{`Env:\gofuck-missing`, `Registry::HKEY_CURRENT_USER\missing`, `C:relative`, `C:\missing*`} {
		script := "Set-Location '" + path + "'"
		output := "Cannot find path '" + path + "' because it does not exist."
		assertNotMatch(t, rule, script, output)
		assertNewCommands(t, rule, script, output, nil)
	}
	assertNewCommands(t, rule, `Set-Location -LiteralPath 'C:\missing''name'`,
		`Cannot find path 'C:\missing'name' because it does not exist.`,
		[]string{`mkdir -Force 'C:\missing''name' && Set-Location -LiteralPath 'C:\missing''name'`})

	assertNewCommands(t, rule, "Set-Location \x27/root/missing\x27", powerShellSetLocationOutput,
		[]string{"mkdir -Force \x27/root/missing\x27 && Set-Location \x27/root/missing\x27"})
	assertNewCommands(t, rule, "pushd \x27/root/missing\x27", powerShellSetLocationOutput,
		[]string{"mkdir -Force \x27/root/missing\x27 && pushd \x27/root/missing\x27"})

	stubPowerShellPathMissing(t, false)
	assertNotMatch(t, rule, "Set-Location \x27/root/missing\x27", powerShellSetLocationOutput)
}
