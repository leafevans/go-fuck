package logs

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/leafevans/go-fuck/v2/internal/consts"
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
// get a compact colored banner; verbose adds the config, instant mode, fzf and
// alias details.
func PrintVersion(info VersionInfo, verbose bool) {
	if !isTerminal(out) {
		Version(info.Version, info.GoVersion, info.Shell)
		return
	}
	fmt.Fprint(out, renderBanner(info, verbose))
}

// OutputIsTerminal reports whether the standard output is a terminal.
//
// Callers can use it to skip work that is only needed for the interactive
// version banner, for example detecting the shell version.
func OutputIsTerminal() bool {
	return isTerminal(out)
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

// logoGradient contains the stops of the horizontal logo gradient.
var logoGradient = []rgb{
	{red: 153, green: 27, blue: 27},  // dark red
	{red: 234, green: 88, blue: 12},  // orange
	{red: 251, green: 191, blue: 36}, // amber
}

// gradientColor returns the color at position t of the logo gradient.
func gradientColor(t float64) rgb {
	t = min(max(t, 0), 1)
	position := t * float64(len(logoGradient)-1)
	index := int(position)
	if index >= len(logoGradient)-1 {
		return logoGradient[len(logoGradient)-1]
	}
	fraction := position - float64(index)
	from, to := logoGradient[index], logoGradient[index+1]
	interpolate := func(a, b int) int {
		return int(math.Round(float64(a) + float64(b-a)*fraction))
	}
	return rgb{
		red:   interpolate(from.red, to.red),
		green: interpolate(from.green, to.green),
		blue:  interpolate(from.blue, to.blue),
	}
}

// renderLogo renders the logo with a horizontal color gradient.
func renderLogo(lines []string, profile colorProfile) string {
	if profile == colorProfileNone {
		return strings.Join(lines, newline) + newline
	}

	width := 0
	for _, line := range lines {
		width = max(width, utf8.RuneCountInString(line))
	}

	var b strings.Builder
	for _, line := range lines {
		if profile == colorProfileANSI {
			b.WriteString(styleBright)
			b.WriteString(foreRed)
			b.WriteString(line)
			b.WriteString(styleReset)
			b.WriteString(newline)
			continue
		}
		b.WriteString(styleBright)
		for index, character := range []rune(line) {
			position := 0.0
			if width > 1 {
				position = float64(index) / float64(width-1)
			}
			b.WriteString(profile.sequence(gradientColor(position), foreRed))
			b.WriteRune(character)
		}
		b.WriteString(styleReset)
		b.WriteString(newline)
	}
	return b.String()
}

// renderBanner renders the colored version banner.
func renderBanner(info VersionInfo, verbose bool) string {
	profile := detectColorProfile()

	logo := []string{
		"   ██████╗  ██████╗     ███████╗██╗   ██╗ ██████╗██╗  ██╗",
		"  ██╔════╝ ██╔═══██╗    ██╔════╝██║   ██║██╔════╝██║ ██╔╝",
		"  ██║  ███╗██║   ██║    █████╗  ██║   ██║██║     █████╔╝",
		"  ██║   ██║██║   ██║    ██╔══╝  ██║   ██║██║     ██╔═██╗",
		"  ╚██████╔╝╚██████╔╝    ██║     ╚██████╔╝╚██████╗██║  ██╗",
		"   ╚═════╝  ╚═════╝     ╚═╝      ╚═════╝  ╚═════╝╚═╝  ╚═╝",
	}

	var b strings.Builder
	b.WriteString(renderLogo(logo, profile))
	b.WriteString(newline)

	b.WriteString(strong(consts.AppName))
	b.WriteString(" ")
	b.WriteString(accent(info.Version))
	b.WriteString(muted("  thefuck in Go"))
	b.WriteString(newline)

	line := func(left, right string) {
		b.WriteString(left)
		b.WriteString(muted("  " + separatorSymbol + "  "))
		b.WriteString(right)
		b.WriteString(newline)
	}
	line("Go "+info.GoVersion, info.Platform)
	line("Shell "+info.Shell, fmt.Sprintf("%d rules", info.Rules))
	line("Home", "https://github.com/leafevans/go-fuck")

	if verbose {
		b.WriteString(newline)
		field := func(label, value string) {
			b.WriteString(muted(fmt.Sprintf("%-10s", label)))
			b.WriteString(value)
			b.WriteString(newline)
		}
		field("Config", info.ConfigPath)
		field("Instant", info.InstantMode)
		if info.Fzf {
			field("fzf", "installed")
		} else {
			field("fzf", "not installed")
		}
		field("Alias", info.Alias)
	}

	return b.String()
}
