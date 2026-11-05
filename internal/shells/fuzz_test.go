package shells

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

// FuzzSplit checks that Split never panics, only returns the documented errors
// and is deterministic. For successful splits it also checks the
// Split -> Quote -> Split round-trip: joining the quoted tokens with spaces
// must split back into exactly the same token list.
//
// The round-trip quotes every token instead of joining the raw tokens because
// a token may contain ASCII whitespace and quotes. Round-tripping the raw input
// string itself is not required: quoting is normalized, so the requirement is
// only that the token list survives, not the original spelling.
func FuzzSplit(f *testing.F) {
	seeds := []string{
		"",
		" ",
		"   ",
		"ls -la",
		"git commit -m 'hello world'",
		"git commit -m \"hello world\"",
		"git commit -m 'fix: don't panic'",
		"echo $PATH && ls ~/file",
		"a\\ b c",
		"a\\\\b",
		"echo \"a \\\"quoted\\\" word\"",
		"printf 'a\tb'",
		"'unterminated",
		"\"unterminated",
		"trailing\\",
		"café 你好 안녕",
		"''",
		"\"\"",
		"a '' b",
		"echo a'b'c",
		"x \\ y",
		"a\\",
		"\\",
		"echo a\\tb",
		"a\t\tb",
		"a\nb",
		"a\rb",
		"git status | grep modified",
		"cmd \"'\"",
		"sed -e 's/\\/foo/bar'",
		"a\\'b",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, command string) {
		args, err := Split(command)
		if err != nil {
			if !errors.Is(err, ErrUnterminatedQuote) && !errors.Is(err, ErrNoEscapedCharacter) {
				t.Fatalf("Split(%q) returned an undocumented error: %v", command, err)
			}
			return
		}

		again, againErr := Split(command)
		if againErr != nil || !slices.Equal(again, args) {
			t.Fatalf("Split(%q) is not deterministic: %q, %v != %q", command, again, againErr, args)
		}

		quoted := make([]string, len(args))
		for index, arg := range args {
			quoted[index] = Quote(arg)
		}
		roundTripped, err := Split(strings.Join(quoted, " "))
		if err != nil {
			t.Fatalf("tokens %q of %q do not quote back into a valid command line: %v", args, command, err)
		}
		if !slices.Equal(roundTripped, args) {
			t.Fatalf("Split(Quote(tokens of %q)) = %q, want %q", command, roundTripped, args)
		}
	})
}
