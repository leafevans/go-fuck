package rules

import "testing"

// dartUnknownCommandTopLevel is the output of the Dart CLI when a top level
// command is misspelled. The message and the suggestion block are produced by
// the args package CommandRunner that the Dart CLI uses.
const dartUnknownCommandTopLevel = `Could not find a command named "buidl".

Did you mean one of these?
  build
  pub
  test

Usage: dart <command|dart-file> [arguments]
Run "dart help <command>" for more information about a command.
`

// dartUnknownCommandSubcommand is the output of a misspelled subcommand.
const dartUnknownCommandSubcommand = `Could not find a subcommand named "upgrad" for "dart pub".

Did you mean one of these?
  upgrade

Usage: dart pub <subcommand> [arguments]
Run "dart help <command>" for more information about a command.
`

// dartUnknownCommandNoSuggestions is printed when no command is close enough to
// be suggested.
const dartUnknownCommandNoSuggestions = `Could not find a command named "zzz".

Usage: dart <command|dart-file> [arguments]
Run "dart help <command>" for more information about a command.
`

func TestRuleDartUnknownCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("dart_unknown_command")
	if rule == nil {
		t.Fatal("dart_unknown_command rule is not registered")
	}

	assertMatch(t, rule, "dart buidl", dartUnknownCommandTopLevel)
	assertMatch(t, rule, "dart pub upgrad", dartUnknownCommandSubcommand)
	assertNotMatch(t, rule, "flutter buidl", dartUnknownCommandTopLevel)
	assertNotMatch(t, rule, "dart pub get", "Resolving dependencies...")
	assertNotMatch(t, rule, "dart zzz", dartUnknownCommandNoSuggestions)

	assertNewCommands(t, rule, "dart buidl", dartUnknownCommandTopLevel,
		[]string{"dart build", "dart pub", "dart test"})
	assertNewCommands(t, rule, "dart pub upgrad", dartUnknownCommandSubcommand,
		[]string{"dart pub upgrade"})
}
