package shells

import (
	"strings"
	"testing"
)

// TestSplitMatchesPythonShlex covers the shlex behaviors that were verified
// against Python 3.12's shlex.split in POSIX mode.
func TestSplitMatchesPythonShlex(t *testing.T) {
	backslash := string(rune(charBackslash))
	doubleQuote := string(rune(charDoubleQuote))
	singleQuote := string(rune(charSingleQuote))
	tab := string(rune(charTab))
	backtick := string(rune(charBacktick))
	nbsp := string(rune(0xA0))

	cases := []struct {
		command  string
		expected []string
	}{
		// Backslashes before characters other than a quote or a backslash are
		// kept inside double quotes.
		{"echo " + doubleQuote + "a " + backslash + "$b" + doubleQuote,
			[]string{"echo", "a " + backslash + "$b"}},
		{"echo " + doubleQuote + "a " + backslash + backtick + "b" + doubleQuote,
			[]string{"echo", "a " + backslash + backtick + "b"}},
		// A backslash before a backslash is removed.
		{"echo " + doubleQuote + "a " + backslash + backslash + "b" + doubleQuote,
			[]string{"echo", "a " + backslash + "b"}},
		// A backslash before a double quote is removed.
		{"echo " + doubleQuote + "a " + backslash + doubleQuote + "b" + doubleQuote,
			[]string{"echo", "a " + doubleQuote + "b"}},
		{"echo " + doubleQuote + backslash + doubleQuote + doubleQuote,
			[]string{"echo", doubleQuote}},
		// A backslash before a normal character is kept.
		{"echo " + doubleQuote + "a" + backslash + "nb" + doubleQuote,
			[]string{"echo", "a" + backslash + "nb"}},
		{"echo " + doubleQuote + "a" + backslash + tab + "b" + doubleQuote,
			[]string{"echo", "a" + backslash + tab + "b"}},
		// Single quotes preserve backslashes.
		{"echo " + singleQuote + "a " + backslash + "$b" + singleQuote,
			[]string{"echo", "a " + backslash + "$b"}},
		// Escaped spaces outside quotes.
		{"a" + backslash + " b", []string{"a b"}},
		{"a" + backslash + backslash + "b", []string{"a" + backslash + "b"}},
		// Unicode whitespace is not a separator.
		{"a" + nbsp + "b", []string{"a" + nbsp + "b"}},
	}
	for _, testCase := range cases {
		got, err := Split(testCase.command)
		if err != nil {
			t.Errorf("Split(%q) error = %v", testCase.command, err)
			continue
		}
		if strings.Join(got, "|") != strings.Join(testCase.expected, "|") {
			t.Errorf("Split(%q) = %q, want %q", testCase.command, got, testCase.expected)
		}
	}

	for _, command := range []string{"a" + backslash, backslash} {
		if _, err := Split(command); err != ErrNoEscapedCharacter {
			t.Errorf("Split(%q) error = %v, want ErrNoEscapedCharacter", command, err)
		}
	}
}
