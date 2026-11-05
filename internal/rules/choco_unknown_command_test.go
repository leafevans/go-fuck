package rules

import (
	"slices"
	"testing"
)

// chocoUnknownCommandHelpOutput is the output of "choco --help" of Chocolatey
// 2.7.3. Every command is listed as a line of the form " * name - description".
const chocoUnknownCommandHelpOutput = "Chocolatey v2.7.3\n" +
	"This is a listing of all of the different things you can pass to choco.\n" +
	"\n" +
	"Options and Switches\n" +
	" -v, --version\n" +
	"     Version - Prints out the Chocolatey version.\n" +
	"\n" +
	"Commands\n" +
	"\n" +
	" * apikey - retrieves or saves an apikey for a particular source\n" +
	" * cache - Manage the local cache\n" +
	" * config - Retrieve and configure config file settings\n" +
	" * export - exports packages to a file (v2.0.0+)\n" +
	" * feature - view and configure choco features\n" +
	" * find - searches remote or local packages\n" +
	" * help - displays top level help information for choco\n" +
	" * info - retrieves package information\n" +
	" * install - installs packages using configured sources\n" +
	" * license - manage your Chocolatey for Business license\n" +
	" * list - lists local packages\n" +
	" * new - generates files necessary for a chocolatey package from a template\n" +
	" * outdated - retrieves packages that are outdated\n" +
	" * pack - packages up a nuspec to a compiled nupkg\n" +
	" * pin - suppress upgrades for a package\n" +
	" * push - pushes a compiled nupkg\n" +
	" * rule - adds or removes package rules (v2.0.0+)\n" +
	" * search - searches remote or local packages\n" +
	" * setapikey - retrieves or saves an apikey for a particular source (alias for apikey)\n" +
	" * source - view and configure default sources\n" +
	" * support - provides support information\n" +
	" * template - generate files for chocolatey packaging\n" +
	" * uninstall - uninstalls a package\n" +
	" * unpackself - have chocolatey set itself up\n" +
	" * upgrade - upgrades packages from various sources\n" +
	"\n" +
	"Please run chocolatey with 'choco command -help' for specific help on\n" +
	" each command.\n" +
	"\n" +
	"How To Pass Options / Switches\n" +
	"\n" +
	" * Unless stated otherwise, an option/switch should only be passed one\n" +
	"   time. Otherwise you may find weird/non-supported behavior.\n" +
	" * Options and switches apply to all items passed, so if you are\n" +
	"   installing multiple packages, and you use --version=1.0.0, choco\n" +
	"   is going to look for and try to install version 1.0.0 of every\n" +
	"   package passed."

// chocoUnknownCommandOutput returns the output Chocolatey prints for an
// unknown command.
func chocoUnknownCommandOutput(name string) string {
	return "Chocolatey v2.7.3" + newline +
		"Could not find a command registered that meets '" + name + "'. Try choco --help for command reference/help." + newline
}

func TestRuleChocoUnknownCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("choco_unknown_command")
	if rule == nil {
		t.Fatal("choco_unknown_command rule is not registered")
	}

	t.Run("parse_commands", func(t *testing.T) {
		got := chocoUnknownCommandParseCommands(chocoUnknownCommandHelpOutput)
		want := append([]string{}, chocoUnknownCommandFallbackCommands...)
		if !slices.Equal(got, want) {
			t.Errorf("parse commands = %q, want %q", got, want)
		}
	})

	originalHelp := chocoUnknownCommandHelp
	t.Cleanup(func() { chocoUnknownCommandHelp = originalHelp })
	chocoUnknownCommandHelp = func() []string {
		return chocoUnknownCommandParseCommands(chocoUnknownCommandHelpOutput)
	}

	assertMatch(t, rule, "choco instal", chocoUnknownCommandOutput("instal"))
	assertNotMatch(t, rule, "choco instal", "")
	assertNotMatch(t, rule, "choco install demo", "Chocolatey installed 1/1 packages."+newline)
	assertNotMatch(t, rule, "cinst instal", chocoUnknownCommandOutput("instal"))

	commandCases := []struct {
		name   string
		script string
		output string
		want   []string
	}{
		{"misspelled", "choco instal", chocoUnknownCommandOutput("instal"), []string{"choco install", "choco uninstall", "choco list"}},
		{"upgrade", "choco upgrad", chocoUnknownCommandOutput("upgrad"), []string{"choco upgrade"}},
		{"unknown", "choco zzz", chocoUnknownCommandOutput("zzz"), []string{}},
	}
	for _, testCase := range commandCases {
		t.Run("new_command/"+testCase.name, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script, testCase.output, testCase.want)
		})
	}
}
