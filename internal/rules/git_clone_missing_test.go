package rules

import "testing"

// gitCloneMissingValidURLs are the URLs that look like the start of a clone.
var gitCloneMissingValidURLs = []string{
	"https://github.com/nvbn/thefuck.git",
	"https://github.com/nvbn/thefuck",
	"http://github.com/nvbn/thefuck.git",
	"git@github.com:nvbn/thefuck.git",
	"git@github.com:nvbn/thefuck",
	"ssh://git@github.com:nvbn/thefuck.git",
}

// gitCloneMissingInvalidURLs are the commands that must not be treated as a
// missing clone.
var gitCloneMissingInvalidURLs = []string{
	"",
	"notacommand",
	"ssh git@github.com:nvbn/thefrick.git",
	"git clone foo",
	"git clone https://github.com/nvbn/thefuck.git",
	"github.com/nvbn/thefuck.git",
	"github.com:nvbn/thefuck.git",
	"git clone git clone ssh://git@github.com:nvbn/thefrick.git",
	"https:/github.com/nvbn/thefuck.git",
}

// gitCloneMissingOutputs are the outputs that report an unknown command.
var gitCloneMissingOutputs = []string{
	"No such file or directory",
	"not found",
	"is not recognised as",
}

func TestRuleGitCloneMissing(t *testing.T) {
	resetTestState(t)
	rule := Get("git_clone_missing")
	if rule == nil {
		t.Fatal("git_clone_missing rule is not registered")
	}

	for _, url := range gitCloneMissingValidURLs {
		for _, output := range gitCloneMissingOutputs {
			t.Run("match/"+url+"/"+output, func(t *testing.T) {
				assertMatch(t, rule, url, output)
			})
		}
	}

	for _, script := range gitCloneMissingInvalidURLs {
		for _, output := range append(append([]string{}, gitCloneMissingOutputs...), "some other output") {
			t.Run("not_match/"+script+"/"+output, func(t *testing.T) {
				assertNotMatch(t, rule, script, output)
			})
		}
	}

	for _, url := range gitCloneMissingValidURLs {
		for _, output := range gitCloneMissingOutputs {
			t.Run("new_command/"+url+"/"+output, func(t *testing.T) {
				assertNewCommands(t, rule, url, output, []string{"git clone " + url})
			})
		}
	}
}
