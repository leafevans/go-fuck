package logs

import (
	"fmt"
	"io"

	"github.com/leafevans/go-fuck/v2/internal/consts"
)

// Status is the shared status vocabulary used by human-readable reports.
type Status string

// The statuses shown by command-line reports.
const (
	StatusOK   Status = "OK"
	StatusInfo Status = "INFO"
	StatusWarn Status = "WARN"
	StatusFail Status = "FAIL"
)

// Renderer styles text for one output stream. Styling is enabled only when the
// stream is a terminal and the user has not disabled colors.
type Renderer struct {
	profile colorProfile
}

// NewRenderer returns a renderer for w.
func NewRenderer(w io.Writer) Renderer {
	profile := detectColorProfile()
	if !IsTerminal(w) {
		profile = colorProfileNone
	}
	return Renderer{profile: profile}
}

// Styled reports whether the renderer emits terminal attributes.
func (r Renderer) Styled() bool {
	return r.profile != colorProfileNone
}

// Strong renders primary text.
func (r Renderer) Strong(text string) string {
	return paint(r.profile, styleBright, text)
}

// Accent renders headings and primary accents.
func (r Renderer) Accent(text string) string {
	return paint(r.profile, r.profile.sequence(accentColor, foreYellow), text)
}

// Muted renders secondary text and hints.
func (r Renderer) Muted(text string) string {
	return paint(r.profile, r.profile.mutedSequence(), text)
}

// Failure renders error text.
func (r Renderer) Failure(text string) string {
	return paint(r.profile, r.profile.sequence(errorColor, foreRed), text)
}

// Title renders the standard command heading.
func (r Renderer) Title(command string) string {
	return r.Strong(consts.ExecutableName()) + " " + r.Accent(command)
}

// Section renders a report section heading.
func (r Renderer) Section(title string) string {
	return r.Accent(title)
}

// Field renders an aligned label followed by its value.
func (r Renderer) Field(label, value string) string {
	return r.Muted(fmt.Sprintf("%-10s", label)) + value
}

// StatusLabel renders an aligned status in the color that matches its meaning.
func (r Renderer) StatusLabel(status Status) string {
	label := fmt.Sprintf("%-4s", status)
	switch status {
	case StatusOK:
		return paint(r.profile, r.profile.sequence(successColor, foreGreen), label)
	case StatusWarn:
		return paint(r.profile, r.profile.sequence(accentColor, foreYellow), label)
	case StatusFail:
		return paint(r.profile, r.profile.sequence(errorColor, foreRed), label)
	default:
		return r.Muted(label)
	}
}

// Diagnostic renders a status and a message on one line.
func (r Renderer) Diagnostic(status Status, message string) string {
	return r.StatusLabel(status) + " " + message
}

// TableRow renders a named doctor-style check.
func (r Renderer) TableRow(name string, status Status, message string) string {
	return fmt.Sprintf("  %-9s %s  %s", name, r.StatusLabel(status), message)
}

// TableHint renders an indented hint below a named check.
func (r Renderer) TableHint(message string) string {
	return fmt.Sprintf("  %-9s %-4s %s", "", "", r.Muted(message))
}
