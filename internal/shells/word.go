package shells

import (
	"regexp"
	"strings"
)

var (
	shellWord      = regexp.MustCompile(`^(?:[^ \t\r\n;&|<>()'"\\]|\\.|'[^']*'|"(?:[^"\\]|\\.)*")+`)
	fishWord       = regexp.MustCompile(`^(?:[^ \t\r\n;&|<>()'"\\]|\\.|'(?:[^'\\]|\\.)*'|"(?:[^"\\]|\\.)*")+`)
	powerShellWord = regexp.MustCompile("^(?:[^ \\t\\r\\n;&|<>()'\"`]|`.|'(?:[^']|'')*'|\"(?:[^\"`]|`.)*\")+")
)

// ReadWord returns the decoded word at the start of rest and its raw byte
// length. It does not evaluate expansions or skip separators.
func ReadWord(rest string, shell Shell) (string, int) {
	for i := range len(rest) {
		switch rest[i] {
		case ' ', '\t', '\r', '\n', ';', '&', '|', '<', '>', '(', ')':
			return rest[:i], i
		case '\'', '"', '\\', '`':
			pattern := shellWord
			switch shell.(type) {
			case *Fish:
				pattern = fishWord
			case *Powershell:
				pattern = powerShellWord
			}
			raw := pattern.FindString(rest)
			if raw == "" || len(raw) < len(rest) && !strings.ContainsRune(" \t\r\n;&|<>()", rune(rest[len(raw)])) {
				return "", 0
			}
			parts := shell.SplitCommand(raw)
			if len(parts) != 1 {
				return "", 0
			}
			return parts[0], len(raw)
		}
	}
	return rest, len(rest)
}
