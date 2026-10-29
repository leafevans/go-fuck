package shells

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/leafevans/go-fuck/internal/consts"
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
func instantModeMarkedAlias(appAlias string) string {
	return "export PS1=\"" + consts.UserCommandMark + "$PS1\";\n" + appAlias
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
