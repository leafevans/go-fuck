package shells

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/consts"
)

// instantModeEnabled reports whether the current shell is the logged shell
// started by the instant mode alias.
func instantModeEnabled() bool {
	for _, name := range []string{consts.InstantModeEnv, consts.LegacyInstantModeEnv} {
		if strings.EqualFold(os.Getenv(name), "true") {
			return true
		}
	}
	return false
}

// instantModeInitAlias returns shell code that starts the logged shell.
//
// The code is evaluated instead of defining the alias: it exports the instant
// mode environment and replaces the current shell with a shell whose terminal
// output is written to a temporary log.
func instantModeInitAlias() string {
	logPath := quoteShellPath(instantModeLogPath())
	return fmt.Sprintf("export %s=True;\nexport %s=%s;\n%s --shell-logger %s;\nrm -f %s;\nexit",
		consts.InstantModeEnv, consts.OutputLogEnv, logPath,
		consts.ExecutableName(), logPath, logPath)
}

// instantModeMarkedAlias prefixes PS1 with the invisible prompt marker so that
// the log reader can split the terminal log into command calls.
//
// The marker is built from an ANSI-C quoted octal literal instead of being
// embedded in the generated code, because the raw zero width characters are not
// portable through every shell configuration file.
//
// The marker is wrapped in the shell's non-printing sequence (bash \[...\],
// zsh %{...%}) so that the prompt width accounting ignores it. A carriage
// return after the marker resets the terminal column in case the terminal
// renders the zero width marker as a cell, which would otherwise indent the
// first prompt line by one column.
//
// hook is shell code that re-adds the marker before every prompt. Prompt
// frameworks such as starship or powerlevel10k rebuild PS1 on every prompt,
// which would otherwise drop the marker.
func instantModeMarkedAlias(appAlias, promptStart, promptEnd, hook string) string {
	return "__gofuck_mark=" + instantModeMarkerLiteral() + ";\n" +
		"export PS1=\"" + promptStart + "${__gofuck_mark}" + promptEnd + "$PS1\";\n" +
		hook + "\n" + appAlias
}

// instantModeMarkerLiteral returns the prompt marker as a POSIX shell ANSI-C
// quoted literal (the zero width user command mark followed by a carriage
// return).
func instantModeMarkerLiteral() string {
	var builder strings.Builder
	builder.WriteString("$'")
	for _, character := range consts.UserCommandMark {
		for _, value := range []byte(string(character)) {
			fmt.Fprintf(&builder, "\\%03o", value)
		}
	}
	builder.WriteString("\\r'")
	return builder.String()
}

// instantModeLogPath returns a fresh temporary path for the terminal log.
func instantModeLogPath() string {
	random := make([]byte, 8)
	if _, err := rand.Read(random); err != nil {
		return filepath.Join(os.TempDir(), fmt.Sprintf("go-fuck-script-log-%d", os.Getpid()))
	}
	return filepath.Join(os.TempDir(), "go-fuck-script-log-"+hex.EncodeToString(random))
}

// quoteShellPath quotes path for POSIX shells.
func quoteShellPath(path string) string {
	const singleQuote = "'"
	const backslash = `\`
	escaped := strings.ReplaceAll(path, singleQuote, singleQuote+backslash+singleQuote+singleQuote)
	return singleQuote + escaped + singleQuote
}
