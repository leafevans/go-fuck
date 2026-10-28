package logs

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/leafevans/go-fuck/internal/consts"
)

// VersionInfo describes the build and environment printed by --version.
type VersionInfo struct {
	// Version is the go-fuck version.
	Version string
	// GoVersion is the Go toolchain version that built the binary.
	GoVersion string
	// Shell is the detected shell, for example "ZSH 5.9".
	Shell string
	// Platform is the operating system and architecture.
	Platform string
	// Rules is the number of compiled in rules.
	Rules int
	// ConfigPath is the user settings file directory.
	ConfigPath string
	// InstantMode describes the instant mode state.
	InstantMode string
	// Fzf reports whether the fzf executable is available.
	Fzf bool
	// Alias is the configured shell alias name.
	Alias string
}

// PrintVersion prints the version banner.
//
// A plain one line version is printed when the output is not a terminal, so
// scripts, pipes and tests keep the previous behavior. Interactive terminals
// get a colored banner with the build and environment details.
func PrintVersion(info VersionInfo) {
	if !isTerminal(out) {
		Version(info.Version, info.GoVersion, info.Shell)
		return
	}
	fmt.Fprint(out, renderBanner(info))
}

// isTerminal reports whether w is a character device such as a terminal.
func isTerminal(w io.Writer) bool {
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// renderBanner renders the colored version banner.
func renderBanner(info VersionInfo) string {
	logo := []string{
		"   ██████╗  ██████╗     ███████╗██╗   ██╗ ██████╗██╗  ██╗",
		"  ██╔════╝ ██╔═══██╗    ██╔════╝██║   ██║██╔════╝██║ ██╔╝",
		"  ██║  ███╗██║   ██║    █████╗  ██║   ██║██║     █████╔╝",
		"  ██║   ██║██║   ██║    ██╔══╝  ██║   ██║██║     ██╔═██╗",
		"  ╚██████╔╝╚██████╔╝    ██║     ╚██████╔╝╚██████╗██║  ██╗",
		"   ╚═════╝  ╚═════╝     ╚═╝      ╚═════╝  ╚═════╝╚═╝  ╚═╝",
	}

	var b strings.Builder
	logoColor := color(styleBright + foreGreen)
	reset := color(styleReset)
	for _, line := range logo {
		b.WriteString(logoColor)
		b.WriteString(line)
		b.WriteString(reset)
		b.WriteString(newline)
	}

	b.WriteString(newline)
	b.WriteString(color(styleBright))
	b.WriteString(consts.AppName)
	b.WriteString(" ")
	b.WriteString(info.Version)
	b.WriteString(reset)
	b.WriteString("  blazingly fast reimplementation of thefuck in Go")
	b.WriteString(newline)
	b.WriteString(color(foreBlue))
	b.WriteString(strings.Repeat("-", 54))
	b.WriteString(reset)
	b.WriteString(newline)

	field := func(label, value string) {
		b.WriteString("  ")
		b.WriteString(color(foreBlue))
		fmt.Fprintf(&b, "%-10s", label)
		b.WriteString(reset)
		b.WriteString(value)
		b.WriteString(newline)
	}

	field("Go", info.GoVersion)
	field("Platform", info.Platform)
	field("Shell", info.Shell)
	field("Rules", fmt.Sprintf("%d", info.Rules))
	field("Config", info.ConfigPath)
	field("Instant", info.InstantMode)
	if info.Fzf {
		field("fzf", "installed")
	} else {
		field("fzf", "not installed")
	}
	field("Alias", info.Alias)
	field("Source", "https://github.com/leafevans/go-fuck")
	b.WriteString(newline)
	return b.String()
}
