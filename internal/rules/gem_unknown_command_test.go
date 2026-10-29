package rules

import (
	"slices"
	"testing"
)

// gemUnknownCommandOutput returns the output gem prints for an unknown
// command.
func gemUnknownCommandOutput(command string) string {
	return "\nERROR:  While executing gem ... (Gem::CommandLineError)\n" +
		"    Unknown command " + command + "\n"
}

const gemUnknownCommandHelpOutput = `
GEM commands are:

    build             Build a gem from a gemspec
    cert              Manage RubyGems certificates and signing settings
    check             Check a gem repository for added or missing files
    cleanup           Clean up old versions of installed gems
    contents          Display the contents of the installed gems
    dependency        Show the dependencies of an installed gem
    environment       Display information about the RubyGems environment
    fetch             Download a gem and place it in the current directory
    generate_index    Generates the index files for a gem server directory
    help              Provide help on the 'gem' command
    install           Install a gem into the local repository
    list              Display local gems whose name matches REGEXP
    lock              Generate a lockdown list of gems
    mirror            Mirror all gem files (requires rubygems-mirror)
    open              Open gem sources in editor
    outdated          Display all gems that need updates
    owner             Manage gem owners of a gem on the push server
    pristine          Restores installed gems to pristine condition from files
                      located in the gem cache
    push              Push a gem up to the gem server
    query             Query gem information in local or remote repositories
    rdoc              Generates RDoc for pre-installed gems
    search            Display remote gems whose name matches REGEXP
    server            Documentation and gem repository HTTP server
    sources           Manage the sources and cache file RubyGems uses to search
                      for gems
    specification     Display gem specification (in yaml)
    stale             List gems along with access times
    uninstall         Uninstall gems from the local repository
    unpack            Unpack an installed gem to the current directory
    update            Update installed gems to the latest version
    which             Find the location of a library file you can require
    yank              Remove a pushed gem from the index

For help on a particular command, use 'gem help COMMAND'.

Commands may be abbreviated, so long as they are unambiguous.
e.g. 'gem i rake' is short for 'gem install rake'.

`

func TestRuleGemUnknownCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("gem_unknown_command")
	if rule == nil {
		t.Fatal("gem_unknown_command rule is not registered")
	}

	originalCommands := gemUnknownCommandGemCommands
	t.Cleanup(func() { gemUnknownCommandGemCommands = originalCommands })
	gemUnknownCommandGemCommands = func() []string {
		return gemUnknownCommandParseCommands(gemUnknownCommandHelpOutput)
	}

	commands := gemUnknownCommandParseCommands(gemUnknownCommandHelpOutput)
	for _, want := range []string{"install", "list"} {
		found := slices.Contains(commands, want)
		if !found {
			t.Errorf("_get_all_commands() = %q, want it to contain %q", commands, want)
		}
	}

	assertMatch(t, rule, "gem isntall jekyll", gemUnknownCommandOutput("isntall"))
	assertMatch(t, rule, "gem last --local", gemUnknownCommandOutput("last"))
	assertNotMatch(t, rule, "gem install jekyll", "")
	assertNotMatch(t, rule, "git log", gemUnknownCommandOutput("log"))

	newCommandCases := []struct {
		script string
		output string
		want   string
	}{
		{"gem isntall jekyll", gemUnknownCommandOutput("isntall"), "gem install jekyll"},
		{"gem last --local", gemUnknownCommandOutput("last"), "gem list --local"},
	}
	for _, testCase := range newCommandCases {
		got, err := rule.GetNewCommand(testCommand(testCase.script, testCase.output))
		if err != nil {
			t.Fatalf("rule %q returned an error: %v", rule.Name, err)
		}
		if len(got) == 0 || got[0] != testCase.want {
			t.Errorf("rule %q returned %q, want first command %q", rule.Name, got, testCase.want)
		}
	}
}
