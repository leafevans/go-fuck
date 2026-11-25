package rules

import (
	"strings"
	"testing"
)

// TestSudoOutputKeywordsCoverPatterns keeps the fast output gate honest: every
// pattern the rule checks must contain one of the keywords, otherwise a long
// capture could skip a pattern that matches.
func TestSudoOutputKeywordsCoverPatterns(t *testing.T) {
	patterns := append(append([]string{}, sudoPatterns...), sudoAuthenticationPatterns...)
	for _, pattern := range patterns {
		covered := false
		for _, keyword := range sudoOutputKeywords {
			if strings.Contains(pattern, keyword) {
				covered = true
				break
			}
		}
		if !covered {
			t.Errorf("pattern %q is not covered by sudoOutputKeywords", pattern)
		}
	}
}

func TestRuleSudo(t *testing.T) {
	resetTestState(t)
	rule := Get("sudo")
	if rule == nil {
		t.Fatal("sudo rule is not registered")
	}

	t.Run("match", func(t *testing.T) {
		for _, output := range []string{
			"Permission denied",
			"permission denied",
			"npm ERR! Error: EACCES, unlink",
			"requested operation requires superuser privilege",
			"need to be root",
			"need root",
			"shutdown: NOT super-user",
			"Error: This command has to be run with superuser privileges (under the root user on most systems).",
			"updatedb: can not open a temporary file for /var/lib/mlocate/mlocate.db",
			"must be root",
			"You don't have access to the history DB.",
			"error: [Errno 13] Permission denied: '/usr/local/lib/python2.7/dist-packages/ipaddr.py'",
		} {
			assertMatch(t, rule, "", output)
		}
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "", "")
		assertNotMatch(t, rule, "sudo ls", "Permission denied")
		// A rejected SSH key is not a missing sudo.
		assertNotMatch(t, rule, "git clone git@github.com:org/private.git",
			"git@github.com: Permission denied (publickey).\nfatal: Could not read from remote repository.")
		assertNotMatch(t, rule, "ssh -T git@github.com", "git@github.com: Permission denied (publickey).")
	})

	t.Run("new command", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			want   []string
		}{
			{"ls", []string{"sudo ls"}},
			{"echo a > b", []string{`sudo sh -c "echo a > b"`}},
			{`echo "a" >> b`, []string{`sudo sh -c "echo \"a\" >> b"`}},
			{"mkdir && touch a", []string{`sudo sh -c "mkdir && touch a"`}},
		} {
			assertNewCommands(t, rule, testCase.script, "", testCase.want)
		}
	})
}
