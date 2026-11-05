package customrules

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/leafevans/go-fuck/internal/consts"
	"github.com/leafevans/go-fuck/internal/types"
)

// fuzzSampleCommands is the fixed set of commands the custom-rule pipeline is
// exercised with. None of the scripts or outputs contains placeholder syntax,
// so fully expanded commands must not contain placeholders either.
var fuzzSampleCommands = []*types.Command{
	types.NewCommand("git remote add origin", "fatal: No such remote 'origin'"),
	types.NewCommand("git remote add upstream", ""),
	types.NewCommandWithoutOutput("git remote add origin"),
	types.NewCommand("echoo hello", "hello"),
}

// FuzzCustomRulePipeline exercises the JSON custom-rule pipeline without
// touching the disk: arbitrary bytes are decoded into the ruleFile layout,
// validated with newDefinition and then matched against the sample commands.
func FuzzCustomRulePipeline(f *testing.F) {
	seeds := []string{
		// Valid rules copied from the package documentation and customrules_test.go.
		`{
  "name": "git_remote_typo",
  "priority": 900,
  "match": {
    "script": "^git remote add (\\w+)$",
    "output": "No such remote '([^']+)'"
  },
  "commands": [
    "git remote remove {1}"
  ]
}`,
		// More valid rules copied from customrules_test.go.
		`{"name": "echo_fix", "match": {"script": "^echoo (.*)$"}, "commands": ["echo {script}"]}`,
		`{"name": "disabled_rule", "enabled": false, "match": {"script": "^x$"}, "commands": ["y"]}`,
		`{"name": "default_priority", "enabled": true, "match": {"output": "(error)"}, "commands": ["echo {1}"]}`,
		`{"name": "all_placeholders", "match": {"script": ".*", "output": "(\\w+)"}, "commands": ["echo {0} {1} {script}"]}`,
		`{"name": "unicode", "match": {"script": "你好"}, "commands": ["echo 世界 {script}"]}`,
		`{"name": "unknown_fields", "match": {"script": ".*"}, "commands": ["echo ok"], "future": [1, 2]}`,
		// Rules copied from the invalid cases of customrules_test.go.
		`{`,
		`{"match": {"script": "^x$"}, "commands": ["y"]}`,
		`{"name": "Bad Name", "match": {"script": "^x$"}, "commands": ["y"]}`,
		`{"name": "no_match", "commands": ["y"]}`,
		`{"name": "no_commands", "match": {"script": "^x$"}}`,
		`{"name": "bad_regex", "match": {"script": "("}, "commands": ["y"]}`,
		`{"name": "bad_priority", "priority": -1, "match": {"script": "^x$"}, "commands": ["y"]}`,
		`{"name": "no_output", "match": {"script": "^x$"}, "commands": ["echo {1}"]}`,
		`{"name": "too_few_groups", "match": {"script": "^x$", "output": "(y)"}, "commands": ["echo {2}"]}`,
		`{"name": "empty_command", "match": {"script": "^x$"}, "commands": [""]}`,
		`null`,
		`[]`,
		`{"name": "null_fields", "enabled": null, "match": {"script": null, "output": null}, "commands": null}`,
		`{"name": "empty", "match": {}, "commands": []}`,
	}
	for _, seed := range seeds {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		var file ruleFile
		if err := json.Unmarshal(data, &file); err != nil {
			return
		}

		definition, err := newDefinition(&file)
		if err != nil {
			if definition != nil {
				t.Fatalf("newDefinition(%q) returned both a definition and an error %v", data, err)
			}
			return
		}
		if definition == nil {
			t.Fatalf("newDefinition(%q) returned a nil definition without an error", data)
		}

		if definition.name != file.Name {
			t.Fatalf("definition name = %q, want %q", definition.name, file.Name)
		}
		wantEnabled := true
		if file.Enabled != nil {
			wantEnabled = *file.Enabled
		}
		if definition.enabled != wantEnabled {
			t.Fatalf("definition enabled = %t, want %t", definition.enabled, wantEnabled)
		}
		wantPriority := file.Priority
		if wantPriority == 0 {
			wantPriority = consts.DefaultPriority
		}
		if definition.priority != wantPriority {
			t.Fatalf("definition priority = %d, want %d", definition.priority, wantPriority)
		}
		if (definition.output != nil) != (file.Match.Output != "") {
			t.Fatalf("definition output presence does not match the rule file: %q", data)
		}

		rule := definition.rule()
		if rule.Name != definition.name ||
			rule.EnabledByDefault != definition.enabled ||
			rule.Priority != definition.priority ||
			rule.RequiresOutput != (definition.output != nil) {
			t.Fatalf("rule() = %s does not mirror the definition for %q", rule, data)
		}

		for _, command := range fuzzSampleCommands {
			matched := definition.matches(command)
			commands, err := definition.newCommands(command)
			if err != nil {
				t.Fatalf("newCommands(%q) error = %v for %q", command.Script, err, data)
			}

			// The pipeline must be deterministic.
			if matchedAgain := definition.matches(command); matchedAgain != matched {
				t.Fatalf("matches(%q) is not deterministic for %q", command.Script, data)
			}
			commandsAgain, againErr := definition.newCommands(command)
			if againErr != nil {
				t.Fatalf("newCommands(%q) is not deterministic for %q: %v", command.Script, data, againErr)
			}
			if !slices.Equal(commands, commandsAgain) {
				t.Fatalf("newCommands(%q) is not deterministic: %q != %q", command.Script, commands, commandsAgain)
			}

			// A rule that requires the output must never match a command
			// whose output was not captured.
			if definition.output != nil && !command.HasOutput() && matched {
				t.Fatalf("rule with match.output matched %q without output", command.Script)
			}
			if rule.IsMatch(command) != matched {
				t.Fatalf("rule.IsMatch(%q) = %t, definition.matches = %t", command.Script, rule.IsMatch(command), matched)
			}
			if !matched {
				continue
			}
			if len(commands) != len(definition.commands) {
				t.Fatalf("newCommands(%q) returned %d commands, want %d", command.Script, len(commands), len(definition.commands))
			}
			for _, corrected := range commands {
				if placeholderPattern.MatchString(corrected) {
					t.Fatalf("expanded command %q still contains a placeholder for %q", corrected, data)
				}
			}
		}
	})
}
