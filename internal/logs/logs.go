// Package logs implements the console output of go-fuck.
//
// All diagnostics are written to stderr so that stdout stays reserved for the
// corrected command. Colors are disabled with the no_colors setting.
package logs

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/consts"
)

// ANSI escape sequences used for colored output.
const (
	styleBright = "\x1b[1m"
	styleReset  = "\x1b[0m"
	foreRed     = "\x1b[31m"
	foreGreen   = "\x1b[32m"
	foreBlue    = "\x1b[34m"
	foreWhite   = "\x1b[37m"
	backRed     = "\x1b[41m"
)

// out is the writer used by all log helpers. It is replaceable in tests.
var out io.Writer = os.Stderr

// newline is the line separator.
var newline = string(rune(10))

// SetOutput redirects log output. It is intended for tests.
func SetOutput(w io.Writer) {
	out = w
}

// color returns the escape sequence unless colors are disabled.
func color(sequence string) string {
	if conf.Current().NoColors {
		return ""
	}
	return sequence
}

// Warn writes a highlighted warning message.
func Warn(title string) {
	fmt.Fprintf(out, "%s[WARN] %s%s\n",
		color(backRed+foreWhite+styleBright), title, color(styleReset))
}

// Exception writes a highlighted warning followed by the error trace.
func Exception(title string, err error) {
	fmt.Fprintf(out, "%s[WARN] %s:%s\n%s%s----------------------------%s\n\n",
		color(backRed+foreWhite+styleBright), title, color(styleReset),
		trace(err), color(backRed+foreWhite+styleBright), color(styleReset))
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
	fmt.Fprintf(out, "%s%s%s\n", color(foreRed), msg, color(styleReset))
}

// NewLine writes a line separator to the log output.
func NewLine() {
	fmt.Fprint(out, newline)
}

// ShowCorrectedCommand prints the selected command without the confirmation
// prompt.
func ShowCorrectedCommand(script string, hasSideEffect bool) {
	fmt.Fprintf(out, "%s%s%s%s%s\n",
		consts.UserCommandMark, color(styleBright), script,
		color(styleReset), sideEffectSuffix(hasSideEffect))
}

// sideEffectSuffix returns the marker appended to commands that change state.
func sideEffectSuffix(hasSideEffect bool) string {
	if hasSideEffect {
		return " (+side effect)"
	}
	return ""
}

// ConfirmText prints the interactive confirmation prompt for a command.
func ConfirmText(script string, hasSideEffect bool) {
	fmt.Fprintf(out, "%s%s%s%s%s%s [%senter%s/%s↑%s/%s↓%s/%sctrl+c%s]",
		consts.UserCommandMark, "\x1b[1K\r", color(styleBright), script,
		color(styleReset), sideEffectSuffix(hasSideEffect),
		color(foreGreen), color(styleReset),
		color(foreBlue), color(styleReset),
		color(foreBlue), color(styleReset),
		color(foreRed), color(styleReset))
}

// Debug writes a debug message when the debug setting is enabled.
func Debug(format string, args ...any) {
	if conf.Current().Debug {
		fmt.Fprintf(out, "%s%sDEBUG:%s %s\n",
			color(foreBlue+styleBright), "", color(styleReset),
			fmt.Sprintf(format, args...))
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

// Version reports the go-fuck, Go and shell versions.
func Version(version, goVersion, shellInfo string) {
	fmt.Fprintf(out, "%s %s using Go %s and %s\n", consts.AppName, version, goVersion, shellInfo)
}
