package logs

import (
	"fmt"
	"os"
	"strings"

	"github.com/leafevans/go-fuck/internal/conf"
)

// colorProfile describes how many colors the terminal can display.
type colorProfile int

// Color profiles, from the least to the most capable terminal.
const (
	colorProfileNone colorProfile = iota
	colorProfileANSI
	colorProfileANSI256
	colorProfileTrueColor
)

// detectColorProfile returns the color capabilities of the terminal.
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

// rgb is a color of the ember theme.
type rgb struct {
	red   int
	green int
	blue  int
}

// The ember theme colors. The accent is the amber of the version banner
// gradient, success and failure keep their usual meaning and muted is used for
// secondary text such as separators and hints.
var (
	accentColor  = rgb{red: 245, green: 158, blue: 11}
	successColor = rgb{red: 34, green: 197, blue: 94}
	errorColor   = rgb{red: 248, green: 113, blue: 113}
	mutedColor   = rgb{red: 148, green: 163, blue: 184}
)

// Theme symbols. They are all part of the basic multilingual plane and are
// covered by the common console fonts.
const (
	selectionSymbol = "›"
	failureSymbol   = "×"
	separatorSymbol = "·"
)

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

// mutedSequence returns the sequence of the muted text color, or an empty
// string when the terminal has no color support.
func (p colorProfile) mutedSequence() string {
	switch p {
	case colorProfileTrueColor:
		return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", mutedColor.red, mutedColor.green, mutedColor.blue)
	case colorProfileANSI256:
		return "\x1b[38;5;245m"
	case colorProfileANSI:
		return "\x1b[90m"
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

// colored wraps text in one of the theme colors.
func colored(color rgb, fallback, text string) string {
	return paint(detectColorProfile(), detectColorProfile().sequence(color, fallback), text)
}

// accent styles the primary accent text.
func accent(text string) string {
	return colored(accentColor, foreYellow, text)
}

// success styles text that confirms an action.
func success(text string) string {
	return colored(successColor, foreGreen, text)
}

// failure styles error text.
func failure(text string) string {
	return colored(errorColor, foreRed, text)
}

// muted styles secondary text.
func muted(text string) string {
	profile := detectColorProfile()
	return paint(profile, profile.mutedSequence(), text)
}

// strong styles the corrected command.
func strong(text string) string {
	return paint(detectColorProfile(), styleBright, text)
}

// Accent returns text styled with the theme accent color.
func Accent(text string) string {
	return accent(text)
}

// Muted returns text styled with the theme muted color.
func Muted(text string) string {
	return muted(text)
}

// Strong returns bold text.
func Strong(text string) string {
	return strong(text)
}

// Success returns text styled with the success color.
func Success(text string) string {
	return success(text)
}

// Failure returns text styled with the failure color.
func Failure(text string) string {
	return failure(text)
}

// ColorsEnabled reports whether the theme emits color sequences.
func ColorsEnabled() bool {
	return detectColorProfile() != colorProfileNone
}

// FailureLine returns a message with the failure symbol, styled as an error
// line. It is used by callers that write directly to standard error.
func FailureLine(message string) string {
	return failure(failureSymbol + " " + message)
}

// WarningLine returns a message with the WARN prefix, styled as a warning line.
// It is used by callers that write directly to standard error.
func WarningLine(message string) string {
	return accent("[WARN]") + " " + message
}
