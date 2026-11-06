package rules

import "testing"

// flutterUnknownCommandTopLevel is the output of the Flutter tool when a
// command is misspelled. FlutterCommandRunner extends the args package
// CommandRunner, so the message and suggestions are the same shape as Dart's.
const flutterUnknownCommandTopLevel = `Could not find a command named "buidl".

Did you mean one of these?
  build
  pub

Usage: flutter <command> [arguments]
Run "flutter help <command>" for more information about a command.
`

func TestRuleFlutterUnknownCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("flutter_unknown_command")
	if rule == nil {
		t.Fatal("flutter_unknown_command rule is not registered")
	}

	assertMatch(t, rule, "flutter buidl", flutterUnknownCommandTopLevel)
	assertNotMatch(t, rule, "dart buidl", flutterUnknownCommandTopLevel)
	assertNotMatch(t, rule, "flutter build", "Building flutter_tools...")
	assertNotMatch(t, rule, "flutter zzz", "Could not find a command named \"zzz\".\n\nUsage: flutter <command>\n")

	assertNewCommands(t, rule, "flutter buidl", flutterUnknownCommandTopLevel,
		[]string{"flutter build", "flutter pub"})
}
