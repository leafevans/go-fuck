package rules

import (
	"slices"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/shells"
)

func TestRuleCommandEditsAcrossShells(t *testing.T) {
	resetTestState(t)
	t.Cleanup(func() { shells.Set(shells.NewGeneric()) })
	stubNodeDirectoryFiles(t, []string{"main file.js"})
	oldGoHelp := goHelpOutput
	goHelpOutput = func(string) string { return "The commands are:\n\n    tidy    update dependencies\n" }
	t.Cleanup(func() { goHelpOutput = oldGoHelp })
	oldNpmHelp := npmWrongCommandHelpCommands
	npmWrongCommandHelpCommands = func() []string { return []string{"install"} }
	t.Cleanup(func() { npmWrongCommandHelpCommands = oldNpmHelp })

	const cargoBin = "error: no bin target named `missing`\navailable bin targets:\n    app\n"
	const cargoBench = "error: no bench target named `missing`\navailable bench targets:\n    benches\n"
	const cargoExample = "error: no example target named `missing`\navailable example targets:\n    sample\n"
	const cargoTest = "error: no test target named `missing`\navailable test targets:\n    integration\n"
	nodeOutput := nodeCannotFindModuleOutput(t, "scripts/mian file.js")
	pipOutput := pipUnknownCommandOutput("instatl", "install")
	for _, shell := range []shells.Shell{shells.NewGeneric(), shells.NewBash(), shells.NewZsh(), shells.NewFish(), shells.NewTcsh(), shells.NewPowershell()} {
		t.Run(shell.Name(), func(t *testing.T) {
			shells.Set(shell)
			for _, tc := range []struct {
				rule, prefix, before, after, suffix, output string
			}{
				{"cargo_no_command", "cargo +stable --config 'foo bar'\t", "'buidl'", "build", " -- buidl", cargoNoCommandModern},
				{"cargo_unknown_command", "cargo --config instll ", "'instll'", "install", "\tinstll", cargoUnknownCommandOutput()},
				{"cargo_missing_bin", "cargo run --package missing --bin ", "'missing'", "app", "\t-- 'missing' --bin missing", cargoBin},
				{"cargo_missing_bench_target", "cargo -v bench --bench=", "'missing'", "benches", " --package 'missing'", cargoBench},
				{"cargo_missing_example_target", "cargo build --manifest-path 'test dir/Cargo.toml' ", "'--example=missing'", "--example=sample", " --release", cargoExample},
				{"cargo_missing_test_target", "cargo +stable --offline test --test ", "'missing'", "integration", " --test other -- --test missing", cargoTest},
				{"cargo_missing_test_target", "cargo test missing --test ", "'missing'", "integration", "", cargoTest},
				{"npm_wrong_command", "sudo npm --prefix 'test dir'\t", "'isntall'", "install", " -- isntall", npmWrongCommandModernOutput},
				{"npm_wrong_command", "npm --prefix=isntall ", "'isntall'", "install", " isntall", npmWrongCommandUsageOutput},
				{"pip_unknown_command", "sudo pip3 --isolated --log instatl ", "'instatl'", "install", " instatl", pipOutput},
				{"yarn_alias", "yarn\t", "'rm'", "remove", " --rm rm", "error Did you mean `yarn remove`?"},
				{"go_unknown_command", "go 'mod'\t", "'tidyy'", "tidy", " -- tidyy", "go mod: unknown command"},
				{"node_script_not_found", "node --no-warnings -- ", "'scripts/mian file.js'", "scripts/main file.js", "\t'scripts/mian file.js'", nodeOutput},
				{"node_script_not_found", "node --env-file 'test dir/.env' ", "'scripts/mian file.js'", "scripts/main file.js", " -- flag", nodeOutput},
			} {
				rule := Get(tc.rule)
				af := shell.Quote(tc.after)
				if value, ok := strings.CutPrefix(tc.after, "--example="); ok {
					af = "--example=" + shell.Quote(value)
				}
				script, want := tc.prefix+tc.before+tc.suffix, tc.prefix+af+tc.suffix
				assertMatch(t, rule, script, tc.output)
				assertNewCommands(t, rule, script, tc.output, []string{want})
				if decoded, size := shells.ReadWord(af, shell); decoded != tc.after || size != len(af) {
					t.Fatalf("replacement %q decodes to %q, want one word %q", af, decoded, tc.after)
				}
				beforeArgs, afterArgs := shell.SplitCommand(script), shell.SplitCommand(want)
				if len(beforeArgs) != len(afterArgs) {
					t.Fatalf("%q changed the argument count: %q -> %q", tc.rule, beforeArgs, afterArgs)
				}
				var changed int
				for i := range beforeArgs {
					if beforeArgs[i] != afterArgs[i] {
						changed++
					}
				}
				if changed != 1 {
					t.Fatalf("%q changed %d arguments: %q -> %q", tc.rule, changed, beforeArgs, afterArgs)
				}
			}

			for _, tc := range []struct{ rule, script, output string }{
				{"cargo_missing_bin", "cargo run --bin missing --bin=missing", cargoBin},
				{"cargo_missing_bin", "cargo run --bin other -- --bin missing", cargoBin},
				{"cargo_missing_bin", "cargo run --bin missing --unsupported value", cargoBin},
				{"cargo_missing_bin", "cargo run --bin -- missing", cargoBin},
				{"cargo_missing_bin", "cargo run --package missing", cargoBin},
				{"cargo_missing_bin", "cargo -- --bin missing", cargoBin},
				{"cargo_missing_bin", "cargo run --bin missing", "error: no bin target named `missing`\navailable bin targets:\n    missing\n"},
				{"cargo_no_command", "cargo build buidl", cargoNoCommandModern},
				{"cargo_no_command", "cargo build buidl", cargoNoCommandUnrecognized},
				{"cargo_no_command", "cargo --config", cargoNoCommandNoSuchSubcommand},
				{"cargo_no_command", "cargo --quiet=true buidl", cargoNoCommandModern},
				{"cargo_unknown_command", "cargo build; cargo instll", cargoUnknownCommandOutput()},
				{"npm_wrong_command", "npm -- isntall", npmWrongCommandModernOutput},
				{"npm_wrong_command", "npm install isntall", npmWrongCommandModernOutput},
				{"pip_unknown_command", "pip --unknown instatl instatl", pipOutput},
				{"pip_unknown_command", "pip --log=instatl list instatl", pipOutput},
				{"pip_unknown_command", "pip instatl 'unfinished", pipOutput},
				{"go_unknown_command", "go mod -- tidyy", "go mod: unknown command"},
				{"go_unknown_command", "go mod mod", "go mod: unknown command"},
				{"yarn_alias", "yarn --cwd rm rm", "error Did you mean `yarn remove`?"},
				{"node_script_not_found", "node -e 'load a module' 'scripts/mian file.js'", nodeOutput},
				{"node_script_not_found", "node 'other/mian file.js' 'scripts/mian file.js'", nodeOutput},
				{"node_script_not_found", "node 'scripts/mian file.js'", nodeOutput + "Require stack:\n- main.js\n"},
				{"node_script_not_found", "node --", nodeOutput},
			} {
				assertNewCommands(t, Get(tc.rule), tc.script, tc.output, nil)
			}
		})
	}
}

func TestSimpleCommandWords(t *testing.T) {
	resetTestState(t)
	if words := simpleCommandWords("npm install # comment"); !slices.Equal(words, []commandWord{{"npm", 0, 3}, {"install", 4, 11}}) {
		t.Fatalf("command words = %v", words)
	}
	for _, script := range []string{"npm install # comment\nother", "npm install | other"} {
		if words := simpleCommandWords(script); words != nil {
			t.Fatalf("simpleCommandWords(%q) = %v", script, words)
		}
	}
}
