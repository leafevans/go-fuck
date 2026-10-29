package shells

import (
	"errors"
	"strings"
)

// Characters used by the POSIX shell parser. Numeric constants are used for
// characters that are awkward to spell in Go string literals.
const (
	charTab            = 9
	charNewline        = 10
	charCarriageReturn = 13
	charSpace          = 32
	charDoubleQuote    = 34
	charSingleQuote    = 39
	charBackslash      = 92
	charBacktick       = 96
)

var (
	// ErrUnterminatedQuote is returned when a command contains an unterminated
	// quote.
	ErrUnterminatedQuote = errors.New("unterminated quote")

	// ErrNoEscapedCharacter is returned when a command ends with a backslash.
	ErrNoEscapedCharacter = errors.New("no escaped character")
)

// isArgumentSpace reports whether ch separates arguments.
//
// Only ASCII spaces are separators to stay compatible with Python's shlex,
// which does not treat Unicode whitespace as a separator.
func isArgumentSpace(ch rune) bool {
	return ch == charSpace || ch == charTab || ch == charCarriageReturn || ch == charNewline
}

// Split splits a command line into arguments using the same rules as Python's
// shlex module in POSIX mode.
//
// Single quotes preserve everything literally. Inside double quotes only a
// backslash before a double quote or another backslash is removed; every other
// backslash sequence is kept. A trailing backslash is an error.
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
		case isArgumentSpace(ch):
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
				if next == charBackslash {
					if i+1 >= len(runes) {
						return nil, ErrUnterminatedQuote
					}
					escaped := runes[i+1]
					if escaped == charDoubleQuote || escaped == charBackslash {
						current.WriteRune(escaped)
					} else {
						current.WriteRune(charBackslash)
						current.WriteRune(escaped)
					}
					i++
					continue
				}
				current.WriteRune(next)
			}
			if !closed {
				return nil, ErrUnterminatedQuote
			}
		case ch == charBackslash:
			if i+1 >= len(runes) {
				return nil, ErrNoEscapedCharacter
			}
			current.WriteRune(runes[i+1])
			inWord = true
			i++
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
