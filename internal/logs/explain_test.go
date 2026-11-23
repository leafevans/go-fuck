package logs

import (
	"strings"
	"testing"
	"time"
)

func TestExplainReportShowsCaptureAndExecutionPlan(t *testing.T) {
	buffer := resetLogsState(t)
	ExplainHeader("git brnch", "instant_log", 2*time.Millisecond)
	ExplainSuggestion(1, "git branch", "git_not_command", "Correct the subcommand", "Refresh cache")
	want := "go-fuck explain\n\n" +
		"Command   git brnch\nCapture   instant_log (2ms)\n\nSuggestions\n" +
		"  1. git branch\n     Rule: git_not_command\n" +
		"     Why: Correct the subcommand\n     Before: Refresh cache\n     Execute: git branch\n"
	if got := buffer.String(); got != want {
		t.Fatalf("explanation = %q, want %q", got, want)
	}
}

func TestExplainReportEscapesTerminalControlsAndOmitsEmptyDetails(t *testing.T) {
	buffer := resetLogsState(t)
	ExplainHeader("echo\nhello", "provided", 0)
	ExplainSuggestion(1, "echo\nhello", "rule\x1b", "", "")
	got := buffer.String()
	if !strings.Contains(got, `echo\nhello`) || !strings.Contains(got, `rule\x1b`) {
		t.Fatalf("explanation did not escape command or rule: %q", got)
	}
	if strings.Contains(got, "Why:") || strings.Contains(got, "Before:") || strings.Contains(got, "\x1b") {
		t.Fatalf("explanation contains empty details or terminal controls: %q", got)
	}
	buffer.Reset()
	ExplainEmpty()
	if got := buffer.String(); got != "  No suggestions\n" {
		t.Fatalf("empty explanation = %q", got)
	}
}
