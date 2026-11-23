package logs

import (
	"strconv"
	"strings"
	"unicode"
)

// CommandText renders a command on one terminal line without interpreting its
// control characters. The original command must still be used for execution.
func CommandText(script string) string {
	index := strings.IndexFunc(script, needsCommandEscape)
	if index < 0 {
		return script
	}
	var text strings.Builder
	text.Grow(len(script))
	text.WriteString(script[:index])
	for _, char := range script[index:] {
		if needsCommandEscape(char) {
			quoted := strconv.QuoteRune(char)
			text.WriteString(quoted[1 : len(quoted)-1])
		} else {
			text.WriteRune(char)
		}
	}
	return text.String()
}

// needsCommandEscape reports whether char could alter terminal layout or text direction.
func needsCommandEscape(char rune) bool {
	return unicode.IsControl(char) || unicode.Is(unicode.Bidi_Control, char) || char == '\u2028' || char == '\u2029'
}

const sideEffectMarker = "(+side effect)"

// CandidateText returns the uncolored label shared by interactive selectors.
func CandidateText(script, reason, beforeRun string) string {
	text := CommandText(script)
	if beforeRun != "" {
		text += "  " + sideEffectMarker
	}
	return text + suggestionDetails(reason, beforeRun)
}

// suggestionDetails formats the optional explanation and side-effect notice.
func suggestionDetails(reason, beforeRun string) string {
	var text string
	if reason != "" {
		text = "  |  " + CommandText(reason)
	}
	if beforeRun != "" {
		text += "  |  Before running: " + CommandText(beforeRun)
	}
	return text
}

// styledSuggestionDetails renders suggestionDetails with muted terminal styling.
func styledSuggestionDetails(reason, beforeRun string) string {
	if details := suggestionDetails(reason, beforeRun); details != "" {
		return muted(details)
	}
	return ""
}
