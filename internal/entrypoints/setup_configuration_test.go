package entrypoints

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSetupIgnoresCommentedConfiguration(t *testing.T) {
	for _, shell := range []string{"bash", "powershell"} {
		t.Run(shell, func(t *testing.T) {
			home := resetEntrypointsState(t)
			t.Setenv("TF_SHELL", shell)
			path := filepath.Join(home, ".bashrc")
			line := `eval "$(go-fuck --alias)"`
			active := `eval "$(go-fuck --alias --capture-mode=auto --enable-experimental-instant-mode)"`
			if shell == "powershell" {
				path = filepath.Join(home, "profile.ps1")
				line = `iex "$(go-fuck --alias)"`
				active = `iex "$(go-fuck --alias --capture-mode=auto --enable-experimental-instant-mode)"`
				if runtime.GOOS != "windows" {
					active = `iex "$(go-fuck --alias --capture-mode=auto)"`
				}
				stubPowerShellProfiles(t, []string{path})
			}
			comment := "# " + line + "\n"
			if err := os.WriteFile(path, []byte(comment), 0o600); err != nil {
				t.Fatal(err)
			}
			for range 2 {
				if code, _, stderr := runMain(t, "go-fuck", "--setup"); code != 0 {
					t.Fatalf("setup failed: %d, %s", code, stderr)
				}
			}
			if got := readFile(t, path); got != comment+active+"\n" {
				t.Fatalf("setup did not append exactly one active line: %q", got)
			}
		})
	}
}

func TestReplaceInstantModeLinePreservesOtherContent(t *testing.T) {
	content := `eval "$(go-fuck --alias --enable-experimental-instant-mode)"`
	legacy := `eval $(go-fuck --alias --enable-experimental-instant-mode)`
	prefix := "# " + legacy + "\r\n" +
		"export OTHER_FLAGS='--enable-experimental-instant-mode'\r\n" +
		"other-tool --enable-experimental-instant-mode\r\n"
	original := prefix + "  " + legacy + "\r\nkeep\r\n"
	path := filepath.Join(t.TempDir(), "config")
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	changed, err := replaceInstantModeLine(path, []byte(original), content)
	if err != nil || !changed {
		t.Fatalf("replaceInstantModeLine() = %t, %v", changed, err)
	}
	want := prefix + "  " + content + "\r\nkeep\r\n"
	if got := readFile(t, path); got != want {
		t.Fatalf("unrelated configuration or line endings changed: %q", got)
	}
	if strings.Count(readFile(t, path), content) != 1 {
		t.Fatal("expected exactly one upgraded alias")
	}
}

func TestHasConfigurationLine(t *testing.T) {
	content := `eval "$(go-fuck --alias)"`
	for _, tc := range []struct {
		name     string
		existing string
		want     bool
	}{
		{"active", content + "\n", true},
		{"indented CRLF", "  \t" + content + "\r\n", true},
		{"inline comment", content + "  # go-fuck\n", true},
		{"BOM", "\uFEFF" + content + "\n", true},
		{"comment", "  # " + content + "\n", false},
		{"other command", "echo '" + content + "'\n", false},
		{"different command", content + "-different\n", false},
		{"empty", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := hasConfigurationLine([]byte(tc.existing), content); got != tc.want {
				t.Fatalf("hasConfigurationLine(%q) = %t, want %t", tc.existing, got, tc.want)
			}
		})
	}
}

func TestReplaceInstantModeLinePreservesBOMAndInlineComment(t *testing.T) {
	content := `eval "$(go-fuck --alias --enable-experimental-instant-mode)"`
	legacy := `eval $(go-fuck --alias --enable-experimental-instant-mode)`
	prefix := "\uFEFF\t"
	suffix := "  # keep this comment\r\n"
	original := prefix + legacy + suffix
	path := filepath.Join(t.TempDir(), "config")
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	changed, err := replaceInstantModeLine(path, []byte(original), content)
	if err != nil || !changed {
		t.Fatalf("replaceInstantModeLine() = %t, %v", changed, err)
	}
	want := prefix + content + suffix
	if got := readFile(t, path); got != want {
		t.Fatalf("BOM, indentation or comment changed: got %q, want %q", got, want)
	}
	if changed, err := replaceInstantModeLine(path, []byte(want), content); err != nil || changed {
		t.Fatalf("repeated replacement = %t, %v", changed, err)
	}
}
