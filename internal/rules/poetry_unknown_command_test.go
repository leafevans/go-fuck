package rules

import "testing"

// poetryUnknownCommandOutput is the output of Poetry for an unknown command.
func poetryUnknownCommandOutput() string {
	return "\nThe requested command intall does not exist.\n\n" +
		"Did you mean one of these perhaps?\n" +
		"    add: Add a new dependency to pyproject.toml\n" +
		"    init: Create a basic pyproject.toml\n" +
		"    install: Installs the project dependencies\n\n" +
		"Documentation: https://python-poetry.org/docs/cli/\n"
}

func TestRulePoetryUnknownCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("poetry_unknown_command")
	if rule == nil {
		t.Fatal("poetry_unknown_command rule is not registered")
	}

	assertMatch(t, rule, "poetry intall", poetryUnknownCommandOutput())
	assertNotMatch(t, rule, "pip intall", poetryUnknownCommandOutput())
	// Without a suggestion list there is nothing to correct.
	assertNotMatch(t, rule, "poetry intall", "The requested command intall does not exist.\n")

	assertNewCommands(t, rule, "poetry intall", poetryUnknownCommandOutput(),
		[]string{"poetry add", "poetry init", "poetry install"})
}

func TestRulePoetryUnknownCommandNamespace(t *testing.T) {
	resetTestState(t)
	rule := Get("poetry_unknown_command")

	// Unknown subcommands inside a namespace use a different message, so the
	// generic rule must not match them.
	output := "The requested command does not exist in the env namespace.\n\n" +
		"Did you mean one of these perhaps?\n" +
		"    env activate: Activate a virtual environment\n"
	assertNotMatch(t, rule, "poetry env actvat", output)
}
