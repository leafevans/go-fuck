// Package logs implements the console output and presentation of go-fuck.
//
// Diagnostics are written to stderr. Command entrypoints write normal
// information to stdout and use NewRenderer so pipes never receive terminal
// escape sequences.
package logs

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/leafevans/go-fuck/v4/internal/conf"
	"github.com/leafevans/go-fuck/v4/internal/consts"
)

// ANSI attributes used by the ember theme. The theme colors themselves are
// defined in theme.go.
const (
	styleBright = "\x1b[1m"
	styleReset  = "\x1b[0m"
	foreRed     = "\x1b[31m"
	foreGreen   = "\x1b[32m"
	foreYellow  = "\x1b[33m"
)

// out is the writer used by all log helpers. It is replaceable in tests.
var out io.Writer = os.Stderr

// newline is the line separator.
var newline = string(rune(10))

// SetOutput redirects log output. It is intended for tests.
func SetOutput(w io.Writer) {
	out = w
}

// Warn writes a highlighted warning message.
func Warn(title string) {
	renderer := NewRenderer(out)
	fmt.Fprintln(out, renderer.Diagnostic(StatusWarn, title))
}

// Exception writes a highlighted warning followed by the error trace.
func Exception(title string, err error) {
	renderer := NewRenderer(out)
	fmt.Fprintf(out, "%s %s:\n%s%s\n\n",
		renderer.StatusLabel(StatusWarn), title, trace(err), renderer.Muted("----------------------------"))
}

// trace renders an error, including its wrapped chain.
func trace(err error) string {
	if err == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%+v\n", err)
}

// RuleFailed reports a rule that returned an error or panicked.
func RuleFailed(ruleName string, err error) {
	Exception("Rule "+ruleName, err)
}

// Failed writes a plain error message.
func Failed(msg string) {
	renderer := NewRenderer(out)
	fmt.Fprintln(out, renderer.Diagnostic(StatusFail, msg))
}

// NewLine writes a line separator to the log output.
func NewLine() {
	fmt.Fprint(out, newline)
}

// sideEffectSuffix returns the marker appended to commands that change state.
func sideEffectSuffix(renderer Renderer, hasSideEffect bool) string {
	if !hasSideEffect {
		return ""
	}
	return "  " + renderer.Accent(sideEffectMarker)
}

// ShowCorrectedCommand prints the selected command without the confirmation
// prompt.
func ShowCorrectedCommand(script string, hasSideEffect bool, reason, beforeRun string) {
	renderer := NewRenderer(out)
	fmt.Fprintf(out, "%s%s %s%s%s\n",
		consts.UserCommandMark, renderer.Accent(selectionSymbol), renderer.Strong(CommandText(script)),
		sideEffectSuffix(renderer, hasSideEffect), styledSuggestionDetails(reason, beforeRun))
}

// ConfirmText prints the interactive confirmation prompt for a command.
//
// The prompt uses the same selection symbol, ember accents and "·" separators
// as the version banner, so all interactive output shares one visual language.
func ConfirmText(script string, hasSideEffect bool, reason, beforeRun string) {
	renderer := NewRenderer(out)
	fmt.Fprintf(out, "%s%s%s %s%s%s  %s%s%s%s%s%s%s",
		consts.UserCommandMark, "\r\x1b[2K", renderer.Accent(selectionSymbol), renderer.Strong(CommandText(script)),
		sideEffectSuffix(renderer, hasSideEffect), styledSuggestionDetails(reason, beforeRun),
		renderer.Muted("["), renderer.Accent("enter"), renderer.Muted(" "+separatorSymbol+" "),
		renderer.Muted("↑↓"), renderer.Muted(" "+separatorSymbol+" "), renderer.Failure("ctrl+c"), renderer.Muted("]"))
}

// PreviewHeader introduces the non-executable preview on the diagnostic stream.
func PreviewHeader() {
	renderer := NewRenderer(out)
	fmt.Fprintf(out, "%s %s\n", renderer.Title("preview"),
		renderer.Muted("original command not re-run; suggestions not applied."))
}

// ExplainHeader starts a non-executing explanation report.
func ExplainHeader(command, source string, elapsed time.Duration) {
	renderer := NewRenderer(out)
	fmt.Fprintln(out, renderer.Title("explain"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, renderer.Field("Command", CommandText(command)))
	fmt.Fprintln(out, renderer.Field("Capture", fmt.Sprintf("%s (%s)", source, elapsed.Round(time.Microsecond))))
	fmt.Fprintln(out)
	fmt.Fprintln(out, renderer.Section("Suggestions"))
}

// ExplainSuggestion prints one suggestion and its non-executing execution plan.
func ExplainSuggestion(index int, script, rule, reason, beforeRun string) {
	renderer := NewRenderer(out)
	fmt.Fprintf(out, "  %d. %s\n", index, renderer.Strong(CommandText(script)))
	fmt.Fprintf(out, "     %s %s\n", renderer.Muted("Rule:"), CommandText(rule))
	if reason != "" {
		fmt.Fprintf(out, "     %s %s\n", renderer.Muted("Why:"), CommandText(reason))
	}
	if beforeRun != "" {
		fmt.Fprintf(out, "     %s %s\n", renderer.Muted("Before:"), CommandText(beforeRun))
	}
	fmt.Fprintf(out, "     %s %s\n", renderer.Muted("Execute:"), CommandText(script))
}

// ExplainEmpty reports that no rule produced a suggestion.
func ExplainEmpty() {
	renderer := NewRenderer(out)
	fmt.Fprintln(out, "  "+renderer.Muted("No suggestions"))
}

// Debug writes a debug message when the debug setting is enabled.
func Debug(format string, args ...any) {
	if conf.Current().Debug {
		renderer := NewRenderer(out)
		fmt.Fprintf(out, "%s %s\n", renderer.Muted("DEBUG"), fmt.Sprintf(format, args...))
	}
}

// noopStop is returned by DebugTime when debug output is disabled.
var noopStop = func() {}

// DebugTime measures the duration of an operation and reports it in debug
// mode. The returned function stops the timer.
//
// A shared no-op is returned when debug output is disabled so that the hot
// rule matching path does not pay for a clock read or a closure allocation.
func DebugTime(message string) func() {
	if !conf.Current().Debug {
		return noopStop
	}
	started := time.Now()
	return func() {
		Debug("%s took: %s", message, time.Since(started))
	}
}
