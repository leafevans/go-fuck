package rules

import (
	"strings"
	"testing"
)

// tsuruNotCommandOutput builds the tsuru error output with the given
// suggestions.
func tsuruNotCommandOutput(suggestions ...string) string {
	var output strings.Builder
	output.WriteString("tsuru: \"tchururu\" is not a tsuru command. See \"tsuru help\"." +
		"\n\nDid you mean?\n")
	for _, suggestion := range suggestions {
		output.WriteString("\t" + suggestion + "\n")
	}
	return output.String()
}

func TestRuleTsuruNotCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("tsuru_not_command")

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "tsuru log", tsuruNotCommandOutput("app-log", "login", "logout"))
		assertMatch(t, rule, "tsuru app-l", tsuruNotCommandOutput("app-list", "app-log"))
		assertMatch(t, rule, "tsuru user-list", tsuruNotCommandOutput("team-user-list"))
		assertMatch(t, rule, "tsuru targetlist", tsuruNotCommandOutput("target-list"))
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "tsuru tchururu",
			"tsuru: \"tchururu\" is not a tsuru command. See \"tsuru help\".\n\nDid you mean?\n")
		assertNotMatch(t, rule, "tsuru version", "tsuru version 0.16.0.")
		assertNotMatch(t, rule, "tsuru help",
			"tsuru version 0.16.0.\n\nUsage: tsuru command [args]\n")
		assertNotMatch(t, rule, "tsuru platform-list",
			"- java\n- logstashgiro\n- newnode\n- nodejs\n- php\n- python\n- python3\n- ruby\n- ruby20\n- static\n")
		assertNotMatch(t, rule, "tsuru env-get", "Error: App thefuck not found.")
	})

	t.Run("get new command", func(t *testing.T) {
		commandOutput := func(broken string, suggestions ...string) string {
			var output strings.Builder
			output.WriteString("tsuru: \"" + broken + "\" is not a tsuru command. See \"tsuru help\"." +
				"\n\nDid you mean?\n")
			for _, suggestion := range suggestions {
				output.WriteString("\t" + suggestion + "\n")
			}
			return output.String()
		}

		assertNewCommands(t, rule, "tsuru log", commandOutput("log", "app-log", "login", "logout"),
			[]string{"tsuru login", "tsuru logout", "tsuru app-log"})
		assertNewCommands(t, rule, "tsuru app-l", commandOutput("app-l", "app-list", "app-log"),
			[]string{"tsuru app-log", "tsuru app-list"})
		assertNewCommands(t, rule, "tsuru user-list", commandOutput("user-list", "team-user-list"),
			[]string{"tsuru team-user-list"})
		assertNewCommands(t, rule, "tsuru targetlist", commandOutput("targetlist", "target-list"),
			[]string{"tsuru target-list"})
	})
}
