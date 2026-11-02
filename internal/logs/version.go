package logs

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/leafevans/go-fuck/internal/conf"
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

// colorProfile describes how many colors the terminal can display.
type colorProfile int

// Color profiles, from the least to the most capable terminal.
const (
	colorProfileNone colorProfile = iota
	colorProfileANSI
	colorProfileANSI256
	colorProfileTrueColor
)

// detectColorProfile returns the color capabilities of the terminal that
// prints the banner.
//
// COLORTERM is set by terminals with 24 bit color support (Windows Terminal,
// iTerm2, Kitty, ...), the "256color" suffix of TERM by 256 color terminals and
// everything else falls back to the basic 16 ANSI colors. The NO_COLOR
// environment variable (https://no-color.org) and the no_colors setting
// disable colors completely.
func detectColorProfile() colorProfile {
	if conf.Current().NoColors || os.Getenv("NO_COLOR") != "" {
		return colorProfileNone
	}
	term := strings.ToLower(os.Getenv("TERM"))
	if term == "dumb" {
		return colorProfileNone
	}
	switch strings.ToLower(os.Getenv("COLORTERM")) {
	case "truecolor", "24bit":
		return colorProfileTrueColor
	}
	if os.Getenv("WT_SESSION") != "" {
		return colorProfileTrueColor
	}
	if strings.Contains(term, "256color") {
		return colorProfileANSI256
	}
	return colorProfileANSI
}

// rgb is a color of the logo gradient.
type rgb struct {
	red   int
	green int
	blue  int
}

// logoGradient contains the stops of the horizontal logo gradient.
var logoGradient = []rgb{
	{red: 239, green: 68, blue: 68},  // red
	{red: 249, green: 115, blue: 22}, // orange
	{red: 250, green: 204, blue: 21}, // gold
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

// ansi256 returns the closest color of the xterm 256 color palette.
func (c rgb) ansi256() int {
	cube := func(value int) int {
		switch {
		case value < 48:
			return 0
		case value < 115:
			return 1
		default:
			return (value - 35) / 40
		}
	}
	return 16 + 36*cube(c.red) + 6*cube(c.green) + cube(c.blue)
}

// sequence returns the ANSI foreground sequence of the color for the profile.
// fallback is used by terminals with the basic 16 colors only.
func (p colorProfile) sequence(c rgb, fallback string) string {
	switch p {
	case colorProfileTrueColor:
		return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", c.red, c.green, c.blue)
	case colorProfileANSI256:
		return fmt.Sprintf("\x1b[38;5;%dm", c.ansi256())
	case colorProfileANSI:
		return fallback
	}
	return ""
}

// muted returns the sequence of the muted text color, or an empty string when
// the terminal has no color support.
func (p colorProfile) muted() string {
	switch p {
	case colorProfileTrueColor:
		return "\x1b[38;2;148;163;184m"
	case colorProfileANSI256:
		return "\x1b[38;5;245m"
	case colorProfileANSI:
		return foreBlue
	}
	return ""
}

// paint wraps text in the sequence when the profile supports colors.
func paint(profile colorProfile, sequence, text string) string {
	if profile == colorProfileNone || sequence == "" {
		return text
	}
	return sequence + text + styleReset
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
	muted := profile.muted()

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

	b.WriteString(paint(profile, styleBright, consts.AppName+" "+info.Version))
	b.WriteString(paint(profile, muted, "  thefuck in Go"))
	b.WriteString(newline)

	line := func(left, right string) {
		b.WriteString(left)
		b.WriteString(paint(profile, muted, "  ·  "))
		b.WriteString(right)
		b.WriteString(newline)
	}
	line("Go "+info.GoVersion, info.Platform)
	line("Shell "+info.Shell, fmt.Sprintf("%d rules", info.Rules))
	line("Home", "https://github.com/leafevans/go-fuck")

	if verbose {
		b.WriteString(newline)
		field := func(label, value string) {
			b.WriteString(paint(profile, muted, fmt.Sprintf("%-10s", label)))
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
