package rules

import "testing"

const powerShellUnknownParameterOutput = "Remove-Item : A parameter cannot be found that matches parameter name 'rf'.\nAt line:1 char:4\n+ rm -rf nosuchdir\n"

const powerShellAmbiguousParameterOutput = "Remove-Item : Parameter cannot be processed because the parameter name 'f' is ambiguous. Possible matches include: -Filter -Force.\n"

func TestRulePowerShellUnixFlags(t *testing.T) {
	resetTestState(t)
	rule := Get("powershell_unix_flags")
	if rule == nil {
		t.Fatal("powershell_unix_flags rule is not registered")
	}

	assertMatch(t, rule, "rm -rf nosuchdir", powerShellUnknownParameterOutput)
	assertMatch(t, rule, "rm -f nosuchdir", powerShellAmbiguousParameterOutput)
	assertMatch(t, rule, "cp -r source dest",
		"Copy-Item : A parameter cannot be found that matches parameter name 'r'.\n")
	assertMatch(t, rule, "mv -f old new", powerShellAmbiguousParameterOutput)

	assertNotMatch(t, rule, "ls -lah", "Get-ChildItem : A parameter cannot be found that matches parameter name 'lah'.\n")
	assertNotMatch(t, rule, "rm -x nosuchdir", "Remove-Item : A parameter cannot be found that matches parameter name 'x'.\n")
	assertNotMatch(t, rule, "rm -rf nosuchdir", "")

	assertNewCommands(t, rule, "rm -rf nosuchdir", powerShellUnknownParameterOutput,
		[]string{"rm -Recurse -Force nosuchdir"})
	assertNewCommands(t, rule, "rm -f nosuchdir", powerShellAmbiguousParameterOutput,
		[]string{"rm -Force nosuchdir"})
	assertNewCommands(t, rule, "cp -R source dest",
		"Copy-Item : A parameter cannot be found that matches parameter name 'R'.\n",
		[]string{"cp -Recurse source dest"})
	assertNewCommands(t, rule, "mv -f old new", powerShellAmbiguousParameterOutput,
		[]string{"mv -Force old new"})
}
