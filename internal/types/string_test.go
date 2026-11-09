package types

import (
	"bytes"
	"strings"
	"testing"
)

func TestStringers(t *testing.T) {
	command := NewCommand("git push", "fatal: no upstream")
	if got := command.String(); !strings.Contains(got, "git push") {
		t.Fatalf("Command.String() = %q", got)
	}
	corrected := &CorrectedCommand{Script: "git push --set-upstream origin main"}
	if got := corrected.String(); !strings.Contains(got, "git push --set-upstream origin main") {
		t.Fatalf("CorrectedCommand.String() = %q", got)
	}
	rule := NewRule("example", func(*Command) bool { return true }, func(*Command) ([]string, error) { return nil, nil })
	rule.Priority = 5
	if got := rule.String(); !strings.Contains(got, "example") || !strings.Contains(got, "priority=5") {
		t.Fatalf("Rule.String() = %q", got)
	}
}

func TestStdoutWriter(t *testing.T) {
	previous := Stdout()
	t.Cleanup(func() { SetStdout(previous) })
	buffer := &bytes.Buffer{}
	SetStdout(buffer)
	if Stdout() != buffer {
		t.Fatalf("Stdout() does not return the writer set with SetStdout")
	}
}
