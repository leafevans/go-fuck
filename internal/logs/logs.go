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

// DebugTime measures the duration of an operation and reports it in debug
// mode. The returned function stops the timer.
func DebugTime(message string) func() {
	started := time.Now()
	return func() {
		Debug("%s took: %s", message, time.Since(started))
	}
}

// HowToConfigureAlias explains how to install the shell alias.
func HowToConfigureAlias(content, path, reload string, canConfigureAutomatically bool) {
	fmt.Fprintf(os.Stdout, "Seems like %sfuck%s alias isn't configured!\n",
		color(styleBright), color(styleReset))

	if content != "" {
		fmt.Fprintf(os.Stdout,
			"Please put %s%s%s in your %s%s%s and apply changes with %s%s%s or restart your shell.\n",
			color(styleBright), content, color(styleReset),
			color(styleBright), path, color(styleReset),
			color(styleBright), reload, color(styleReset))

		if canConfigureAutomatically {
			fmt.Fprintf(os.Stdout,
				"Or run %sfuck%s a second time to configure it automatically.\n",
				color(styleBright), color(styleReset))
		}
	}

	fmt.Fprintln(os.Stdout, "More details - https://github.com/leafevans/go-fuck#manual-installation")
}

// AlreadyConfigured reports that the alias is already present in the shell
// configuration.
func AlreadyConfigured(reload string) {
	fmt.Fprintf(os.Stdout,
		"Seems like %sfuck%s alias already configured!\nFor applying changes run %s%s%s or restart your shell.\n",
		color(styleBright), color(styleReset),
		color(styleBright), reload, color(styleReset))
}

// ConfiguredSuccessfully reports that the alias was added to the shell
// configuration.
func ConfiguredSuccessfully(reload string) {
	fmt.Fprintf(os.Stdout,
		"%sfuck%s alias configured successfully!\nFor applying changes run %s%s%s or restart your shell.\n",
		color(styleBright), color(styleReset),
		color(styleBright), reload, color(styleReset))
}

// Version reports the go-fuck, Go and shell versions.
func Version(version, goVersion, shellInfo string) {
	fmt.Fprintf(out, "%s %s using Go %s and %s\n", consts.AppName, version, goVersion, shellInfo)
}
