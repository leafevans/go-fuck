package argumentparser

import (
	"strings"
	"testing"
)

func TestParseEmpty(t *testing.T) {
	args, err := Parse([]string{"go-fuck"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(args.Command) != 0 || args.AliasRequested || args.Version || args.Help {
		t.Fatalf("unexpected args: %+v", args)
	}
}

func TestParseCommandWithoutPlaceholder(t *testing.T) {
	args, err := Parse([]string{"go-fuck", "git", "brnch"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if strings.Join(args.Command, " ") != "git brnch" {
		t.Fatalf("Command = %q, want %q", args.Command, "git brnch")
	}
}

func TestParsePlaceholder(t *testing.T) {
	// The shell alias calls "go-fuck PLACEHOLDER $@", where $@ contains the
	// flags the user passed to the alias.
	args, err := Parse([]string{"go-fuck", "GOFUCK_ARGUMENT_PLACEHOLDER", "-y", "-d"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if !args.Yes || !args.Debug {
		t.Fatalf("flags after the placeholder were not parsed: %+v", args)
	}
	if len(args.Command) != 0 {
		t.Fatalf("Command = %q, want no command", args.Command)
	}
}

func TestParseLegacyPlaceholder(t *testing.T) {
	args, err := Parse([]string{"go-fuck", "THEFUCK_ARGUMENT_PLACEHOLDER", "ls"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if strings.Join(args.Command, " ") != "ls" {
		t.Fatalf("Command = %q, want %q", args.Command, "ls")
	}
}

func TestParseFlags(t *testing.T) {
	args, err := Parse([]string{"go-fuck", "-v", "-d", "-y", "--enable-experimental-instant-mode"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if !args.Version || !args.Debug || !args.Yes || !args.EnableExperimentalInstantMode {
		t.Fatalf("unexpected args: %+v", args)
	}
}

func TestParseHelpAndVersion(t *testing.T) {
	args, err := Parse([]string{"go-fuck", "-h"})
	if err != nil || !args.Help {
		t.Fatalf("Parse(-h) = %+v, %v", args, err)
	}
	args, err = Parse([]string{"go-fuck", "--version"})
	if err != nil || !args.Version {
		t.Fatalf("Parse(--version) = %+v, %v", args, err)
	}
}

func TestParseSetup(t *testing.T) {
	args, err := Parse([]string{"go-fuck", "--setup"})
	if err != nil || !args.Setup {
		t.Fatalf("Parse(--setup) = %+v, %v", args, err)
	}
}

func TestParseRepeat(t *testing.T) {
	args, err := Parse([]string{"go-fuck", "-r"})
	if err != nil || !args.Repeat {
		t.Fatalf("Parse(-r) = %+v, %v", args, err)
	}
}

func TestParseYesAndRepeatConflict(t *testing.T) {
	if _, err := Parse([]string{"go-fuck", "-y", "-r"}); err == nil {
		t.Fatal("Parse(-y -r) should fail")
	}
}

func TestParseAlias(t *testing.T) {
	t.Setenv("TF_ALIAS", "fuck")
	args, err := Parse([]string{"go-fuck", "-a"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if !args.AliasRequested || args.Alias != "fuck" {
		t.Fatalf("unexpected args: %+v", args)
	}

	args, err = Parse([]string{"go-fuck", "-a", "my-fuck"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if !args.AliasRequested || args.Alias != "my-fuck" {
		t.Fatalf("unexpected args: %+v", args)
	}

	args, err = Parse([]string{"go-fuck", "--alias=other"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if !args.AliasRequested || args.Alias != "other" {
		t.Fatalf("unexpected args: %+v", args)
	}
}

func TestParseValueArguments(t *testing.T) {
	args, err := Parse([]string{"go-fuck", "-l", "/tmp/log", "--force-command", "git brnch"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if args.ShellLogger != "/tmp/log" {
		t.Fatalf("ShellLogger = %q", args.ShellLogger)
	}
	if args.ForceCommand != "git brnch" {
		t.Fatalf("ForceCommand = %q", args.ForceCommand)
	}

	args, err = Parse([]string{"go-fuck", "--shell-logger=log", "--force-command=ls"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if args.ShellLogger != "log" || args.ForceCommand != "ls" {
		t.Fatalf("unexpected args: %+v", args)
	}
}

func TestParseMissingValue(t *testing.T) {
	for _, argv := range [][]string{
		{"go-fuck", "-l"},
		{"go-fuck", "--force-command"},
	} {
		if _, err := Parse(argv); err == nil {
			t.Errorf("Parse(%q) should fail", argv)
		}
	}
}

func TestParseUnknownFlag(t *testing.T) {
	if _, err := Parse([]string{"go-fuck", "--unknown"}); err == nil {
		t.Fatal("Parse(--unknown) should fail")
	}
}

func TestParseDoubleDash(t *testing.T) {
	args, err := Parse([]string{"go-fuck", "--", "-d", "ls"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if args.Debug {
		t.Fatal("-d after -- was parsed as a flag")
	}
	if strings.Join(args.Command, " ") != "-d ls" {
		t.Fatalf("Command = %q, want %q", args.Command, "-d ls")
	}
}
