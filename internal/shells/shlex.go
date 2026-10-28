package shells

import (
	"errors"
	"strings"
	"unicode"
)

// Characters used by the POSIX shell parser. Numeric constants are used for
// characters that are awkward to spell in Go string literals.
const (
	charDoubleQuote = 34
	charSingleQuote = 39
	charDollar      = 36
	charBackslash   = 92
	charBacktick    = 96
	charNewline     = 10
)

// ErrUnterminatedQuote is returned when a command contains an unterminated
// quote.
var ErrUnterminatedQuote = errors.New("unterminated quote")

// Split splits a command line into arguments using POSIX shell rules.
//
// Single quotes preserve everything literally. Double quotes preserve
// everything except backslash escapes, and unquoted backslashes escape the next
// character.
func Split(command string) ([]string, error) {
	var (
		args    []string
		current strings.Builder
		inWord  bool
	)

	runes := []rune(command)
	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		switch {
		case unicode.IsSpace(ch):
			if inWord {
				args = append(args, current.String())
				current.Reset()
				inWord = false
			}
		case ch == charSingleQuote:
			inWord = true
			closed := false
			for i++; i < len(runes); i++ {
				if runes[i] == charSingleQuote {
					closed = true
					break
				}
				current.WriteRune(runes[i])
			}
			if !closed {
				return nil, ErrUnterminatedQuote
			}
		case ch == charDoubleQuote:
			inWord = true
			closed := false
			for i++; i < len(runes); i++ {
				next := runes[i]
				if next == charDoubleQuote {
					closed = true
					break
				}
				if next == charBackslash && i+1 < len(runes) {
					escaped := runes[i+1]
					switch escaped {
					case charDoubleQuote, charBackslash, charDollar, charBacktick, charNewline:
						current.WriteRune(escaped)
						i++
						continue
					}
				}
				current.WriteRune(next)
			}
			if !closed {
				return nil, ErrUnterminatedQuote
			}
		case ch == charBackslash:
			if i+1 < len(runes) {
				current.WriteRune(runes[i+1])
				inWord = true
				i++
			}
		default:
			current.WriteRune(ch)
			inWord = true
		}
	}
	if inWord {
		args = append(args, current.String())
	}
	return args, nil
}

// Quote returns a POSIX shell escaped version of s, mirroring shlex.quote.
func Quote(s string) string {
	if s == "" {
		return "''"
	}
	if isSafeWord(s) {
		return s
	}
	doubleQuote := string(rune(charDoubleQuote))
	escaped := "'" + doubleQuote + "'" + doubleQuote + "'"
	return "'" + strings.ReplaceAll(s, "'", escaped) + "'"
}

// isSafeWord reports whether s can be used unquoted in a shell command.
func isSafeWord(s string) bool {
	for _, ch := range s {
		switch {
		case ch >= 'a' && ch <= 'z',
			ch >= 'A' && ch <= 'Z',
			ch >= '0' && ch <= '9',
			strings.ContainsRune("_@%+=:,./-", ch):
		default:
			return false
		}
	}
	return true
}
