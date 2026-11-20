package logs

import (
	"strconv"
	"strings"
	"unicode"
)

// CommandText renders a command on one terminal line without interpreting its
// control characters. The original command must still be used for execution.
func CommandText(script string) string {
	index := strings.IndexFunc(script, unicode.IsControl)
	if index < 0 {
		return script
	}
	var text strings.Builder
	text.Grow(len(script))
	text.WriteString(script[:index])
	for _, char := range script[index:] {
		if unicode.IsControl(char) {
			quoted := strconv.QuoteRune(char)
			text.WriteString(quoted[1 : len(quoted)-1])
		} else {
			text.WriteRune(char)
		}
	}
	return text.String()
}
