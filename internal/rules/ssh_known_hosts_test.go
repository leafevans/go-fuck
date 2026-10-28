package rules

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// sshKnownHostsFixture creates a known_hosts file and the ssh error output
// that points at its second line.
func sshKnownHostsFixture(t *testing.T) (errorOutput, path string) {
	t.Helper()
	path = filepath.Join(t.TempDir(), "known_hosts")
	content := strings.Join([]string{
		"123.234.567.890 asdjkasjdakjsd",
		"98.765.432.321 ejioweojwejrosj",
		"111.222.333.444 qwepoiwqepoiss",
		"",
	}, "\n")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("create known_hosts: %v", err)
	}

	errorOutput = strings.Join([]string{
		"@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@",
		"@    WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED!     @",
		"@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@",
		"IT IS POSSIBLE THAT SOMEONE IS DOING SOMETHING NASTY!",
		"Someone could be eavesdropping on you right now (man-in-the-middle attack)!",
		"It is also possible that a host key has just been changed.",
		"The fingerprint for the RSA key sent by the remote host is",
		"b6:cb:07:34:c0:a0:94:d3:0d:69:83:31:f4:c5:20:9b.",
		"Please contact your system administrator.",
		"Add correct host key in " + path + " to get rid of this message.",
		"Offending RSA key in " + path + ":2",
		"RSA host key for 98.765.432.321 has changed and you have requested strict checking.",
		"Host key verification failed.",
	}, "\n")
	return errorOutput, path
}

func TestRuleSSHKnownHosts(t *testing.T) {
	resetTestState(t)
	rule := Get("ssh_known_hosts")
	if rule == nil {
		t.Fatal("ssh_known_hosts rule is not registered")
	}

	errorOutput, path := sshKnownHostsFixture(t)

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "ssh", errorOutput)
		assertMatch(t, rule, "ssh user@host", errorOutput)
		assertMatch(t, rule, "scp something something", errorOutput)
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, errorOutput, "")
		assertNotMatch(t, rule, "notssh", errorOutput)
		assertNotMatch(t, rule, "ssh", "")
	})

	t.Run("new command", func(t *testing.T) {
		assertNewCommands(t, rule, "ssh user@host", errorOutput,
			[]string{"ssh user@host"})
	})

	t.Run("side effect", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping on Windows like the original test")
		}
		command := testCommand("ssh user@host", errorOutput)
		if err := rule.SideEffect(command, ""); err != nil {
			t.Fatalf("side effect returned an error: %v", err)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read known_hosts: %v", err)
		}
		want := "123.234.567.890 asdjkasjdakjsd\n111.222.333.444 qwepoiwqepoiss\n"
		if string(content) != want {
			t.Fatalf("side effect left %q, want %q", string(content), want)
		}
	})
}
