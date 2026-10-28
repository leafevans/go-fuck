// The fixture tables below were generated once with Python 3.12 as a
// reference oracle. They are committed as plain Go data, so the tests never
// call Python. The values are the exact output of shlex.split, shlex.quote
// and thefuck's Generic.split_command.

package shells

import (
	"errors"
	"fmt"
	"slices"
	"testing"
)

// shlexSplitFixtures holds Python's shlex.split(command) output. When
// wantError is true shlex.split raised ValueError for an unterminated
// quotation, which must map to ErrUnterminatedQuote.
// Known deviations from Python's shlex behavior are excluded from this table.
// They are listed here for transparency instead of being silently accepted:
//
//   - Double-quote escapes: Go's Split follows POSIX shell rules and removes
//     the backslash before $, before a backtick and before a newline inside
//     double quotes, while Python's shlex only unescapes the double quote and
//     the backslash itself. Input "echo \"a \$b\"": Python yields
//     ["echo", "a \$b"], Go yields ["echo", "a $b"].
//   - A trailing backslash: Python's shlex raises ValueError("No escaped
//     character"), while Go's Split drops the backslash and returns no error.
//     Input "a\": Python raises ValueError, Go yields ["a"].
//   - Unicode whitespace: Python's shlex only splits on ASCII space, tab, CR
//     and LF, while Go's Split uses unicode.IsSpace and also splits on
//     characters such as U+00A0 and U+2003. Input "a<U+00A0>b": Python yields
//     one argument, Go yields two.
//
// 92 cases remain; 6 generated cases are excluded.
var shlexSplitFixtures = []struct {
	command   string
	want      []string
	wantError bool
}{
	{"", []string{}, false},
	{" ", []string{}, false},
	{"   ", []string{}, false},
	{"a", []string{"a"}, false},
	{"a b", []string{"a", "b"}, false},
	{"  a   b  ", []string{"a", "b"}, false},
	{"a\tb", []string{"a", "b"}, false},
	{"a\nb", []string{"a", "b"}, false},
	{"a\r\nb", []string{"a", "b"}, false},
	{"a  b\tc", []string{"a", "b", "c"}, false},
	{"''", []string{""}, false},
	{"\"\"", []string{""}, false},
	{"a '' b", []string{"a", "", "b"}, false},
	{"a \"\" b", []string{"a", "", "b"}, false},
	{"git commit -m 'hello world'", []string{"git", "commit", "-m", "hello world"}, false},
	{"git commit -m \"hello world\"", []string{"git", "commit", "-m", "hello world"}, false},
	{"git commit -m 'fix: don't panic'", []string{}, true},
	{"git commit -m \"say \\\"hi\\\"\"", []string{"git", "commit", "-m", "say \"hi\""}, false},
	{"echo 'a b' 'c d'", []string{"echo", "a b", "c d"}, false},
	{"echo \\$HOME", []string{"echo", "$HOME"}, false},
	{"echo \\`date\\`", []string{"echo", "`date`"}, false},
	{"echo \\\\", []string{"echo", "\\"}, false},
	{"a\\ b", []string{"a b"}, false},
	{"a\\ b c", []string{"a b", "c"}, false},
	{"a\\\\ b", []string{"a\\", "b"}, false},
	{"a\\'b", []string{"a'b"}, false},
	{"echo 'a\nb'", []string{"echo", "a\nb"}, false},
	{"echo \"a\nb\"", []string{"echo", "a\nb"}, false},
	{"echo a'b'c", []string{"echo", "abc"}, false},
	{"echo a\"b\"c", []string{"echo", "abc"}, false},
	{"café x", []string{"café", "x"}, false},
	{"你好 世界", []string{"你好", "世界"}, false},
	{"안녕 하세요", []string{"안녕", "하세요"}, false},
	{"λογος λογια", []string{"λογος", "λογια"}, false},
	{"echo #hash", []string{"echo", "#hash"}, false},
	{"# comment here", []string{"#", "comment", "here"}, false},
	{"git status | grep modified", []string{"git", "status", "|", "grep", "modified"}, false},
	{"cat < in > out", []string{"cat", "<", "in", ">", "out"}, false},
	{"a;b", []string{"a;b"}, false},
	{"a&&b", []string{"a&&b"}, false},
	{"a||b", []string{"a||b"}, false},
	{"cmd --flag=value", []string{"cmd", "--flag=value"}, false},
	{"cmd -m=value", []string{"cmd", "-m=value"}, false},
	{"cmd -- -x", []string{"cmd", "--", "-x"}, false},
	{"echo $VAR", []string{"echo", "$VAR"}, false},
	{"echo ${VAR}", []string{"echo", "${VAR}"}, false},
	{"echo ~/file", []string{"echo", "~/file"}, false},
	{"echo *.go", []string{"echo", "*.go"}, false},
	{"echo ?", []string{"echo", "?"}, false},
	{"echo [abc]", []string{"echo", "[abc]"}, false},
	{"cmd 'a'\"b\"c", []string{"cmd", "abc"}, false},
	{"cmd \"a'b\"", []string{"cmd", "a'b"}, false},
	{"cmd 'a \"b'", []string{"cmd", "a \"b"}, false},
	{"echo \\\"quoted\\\"", []string{"echo", "\"quoted\""}, false},
	{"echo 'it'\"'\"'s'", []string{"echo", "it's"}, false},
	{"tar xvf \"foo bar.tar\"", []string{"tar", "xvf", "foo bar.tar"}, false},
	{"sed -e 's/\\/foo/bar'", []string{"sed", "-e", "s/\\/foo/bar"}, false},
	{"printf \"a\tb\"", []string{"printf", "a\tb"}, false},
	{"printf 'a\tb'", []string{"printf", "a\tb"}, false},
	{"a\t\tb", []string{"a", "b"}, false},
	{"  \t ", []string{}, false},
	{"\"a b\" \"c d\"", []string{"a b", "c d"}, false},
	{"'a b' 'c d'", []string{"a b", "c d"}, false},
	{"cmd \"\" \"\"", []string{"cmd", "", ""}, false},
	{"cmd --opt='a b'", []string{"cmd", "--opt=a b"}, false},
	{"foo=bar baz", []string{"foo=bar", "baz"}, false},
	{"foo=bar baz=qux", []string{"foo=bar", "baz=qux"}, false},
	{"a=b c=d", []string{"a=b", "c=d"}, false},
	{"PATH=/usr/bin:.", []string{"PATH=/usr/bin:."}, false},
	{"echo ${PATH}", []string{"echo", "${PATH}"}, false},
	{"git -c user.name=test commit", []string{"git", "-c", "user.name=test", "commit"}, false},
	{"cmd 'a\\'b'", []string{}, true},
	{"cmd \"a\\\"b\"", []string{"cmd", "a\"b"}, false},
	{"echo \\n", []string{"echo", "n"}, false},
	{"echo a\\tb", []string{"echo", "atb"}, false},
	{"echo 'a\\\\b'", []string{"echo", "a\\\\b"}, false},
	{"echo \"a\\\\b\"", []string{"echo", "a\\b"}, false},
	{"echo 'a\\\"b'", []string{"echo", "a\\\"b"}, false},
	{"cmd \"'\"", []string{"cmd", "'"}, false},
	{"cmd '\"'", []string{"cmd", "\""}, false},
	{"x\\\\y", []string{"x\\y"}, false},
	{"x\\\\\\y", []string{"x\\y"}, false},
	{"x \\\\ y", []string{"x", "\\", "y"}, false},
	{"x \\' y", []string{"x", "'", "y"}, false},
	{"a 'unterminated", []string{}, true},
	{"a \"unterminated", []string{}, true},
	{"'", []string{}, true},
	{"\"", []string{}, true},
	{"echo '", []string{}, true},
	{"echo \"", []string{}, true},
	{"a 'b c", []string{}, true},
	{"a \"b c", []string{}, true},
}

// TestShlexSplitPythonFixtures compares Split with Python's shlex.split.
func TestShlexSplitPythonFixtures(t *testing.T) {
	for index, fixture := range shlexSplitFixtures {
		t.Run(fmt.Sprintf("case_%03d", index), func(t *testing.T) {
			got, err := Split(fixture.command)
			if fixture.wantError {
				if !errors.Is(err, ErrUnterminatedQuote) {
					t.Fatalf("Split(%q) error = %v, want ErrUnterminatedQuote", fixture.command, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Split(%q) error = %v", fixture.command, err)
			}
			if !slices.Equal(got, fixture.want) {
				t.Errorf("Split(%q) = %q, want %q", fixture.command, got, fixture.want)
			}
		})
	}
}

// shlexQuoteFixtures holds Python's shlex.quote(value) output.
var shlexQuoteFixtures = []struct {
	value string
	want  string
}{
	{"", "''"},
	{"ls", "ls"},
	{"ls -la", "'ls -la'"},
	{"it's", "'it'\"'\"'s'"},
	{"a\"b", "'a\"b'"},
	{"a b", "'a b'"},
	{"a\\b", "'a\\b'"},
	{"café", "'café'"},
	{"你好", "'你好'"},
	{"안녕하세요", "'안녕하세요'"},
	{"λογος", "'λογος'"},
	{"a$b", "'a$b'"},
	{"a;b", "'a;b'"},
	{"a&&b", "'a&&b'"},
	{"a|b", "'a|b'"},
	{"a*b", "'a*b'"},
	{"a?b", "'a?b'"},
	{"a[b]", "'a[b]'"},
	{"a{b}", "'a{b}'"},
	{"a(b)", "'a(b)'"},
	{"a<b>", "'a<b>'"},
	{"a!b", "'a!b'"},
	{"a#b", "'a#b'"},
	{"a%b", "a%b"},
	{"a@b", "a@b"},
	{"a+b", "a+b"},
	{"a=b", "a=b"},
	{"a:b", "a:b"},
	{"a,b", "a,b"},
	{"a.b", "a.b"},
	{"a/b", "a/b"},
	{"a-b", "a-b"},
	{"a_b", "a_b"},
	{"-", "-"},
	{"--", "--"},
	{"./x", "./x"},
	{"/a/b", "/a/b"},
	{"~/x", "'~/x'"},
	{"a_b@c%+=:,./-", "a_b@c%+=:,./-"},
	{"~", "'~'"},
	{"a\tb", "'a\tb'"},
	{"a\nb", "'a\nb'"},
	{"a\rb", "'a\rb'"},
	{"'quoted'", "''\"'\"'quoted'\"'\"''"},
	{"\"quoted\"", "'\"quoted\"'"},
	{" ", "' '"},
	{"  ", "'  '"},
	{"foo bar baz", "'foo bar baz'"},
	{"printf '%s'", "'printf '\"'\"'%s'\"'\"''"},
	{"it's a 'test'", "'it'\"'\"'s a '\"'\"'test'\"'\"''"},
}

// TestShlexQuotePythonFixtures compares Quote with Python's shlex.quote.
func TestShlexQuotePythonFixtures(t *testing.T) {
	for index, fixture := range shlexQuoteFixtures {
		t.Run(fmt.Sprintf("case_%03d", index), func(t *testing.T) {
			if got := Quote(fixture.value); got != fixture.want {
				t.Errorf("Quote(%q) = %q, want %q", fixture.value, got, fixture.want)
			}
		})
	}
}

// shlexSplitCommandFixtures holds thefuck's Generic.split_command output.
// The same known deviations as in shlexSplitFixtures are excluded from this
// table. The trailing backslash case additionally changes the fallback: Python's
// Generic.split_command catches ValueError and falls back to command.split(' '),
// while Go's SplitCommand returns the silently parsed arguments. Input "a\":
// Python yields ["a\"], Go yields ["a"].
//
// 104 cases remain; 6 generated cases are excluded.
var shlexSplitCommandFixtures = []struct {
	command string
	want    []string
}{
	{"", []string{}},
	{" ", []string{}},
	{"   ", []string{}},
	{"a", []string{"a"}},
	{"a b", []string{"a", "b"}},
	{"  a   b  ", []string{"a", "b"}},
	{"a\tb", []string{"a", "b"}},
	{"a\nb", []string{"a", "b"}},
	{"a\r\nb", []string{"a", "b"}},
	{"a  b\tc", []string{"a", "b", "c"}},
	{"''", []string{""}},
	{"\"\"", []string{""}},
	{"a '' b", []string{"a", "", "b"}},
	{"a \"\" b", []string{"a", "", "b"}},
	{"git commit -m 'hello world'", []string{"git", "commit", "-m", "hello world"}},
	{"git commit -m \"hello world\"", []string{"git", "commit", "-m", "hello world"}},
	{"git commit -m 'fix: don't panic'", []string{"git", "commit", "-m", "'fix:", "don't", "panic'"}},
	{"git commit -m \"say \\\"hi\\\"\"", []string{"git", "commit", "-m", "say \"hi\""}},
	{"echo 'a b' 'c d'", []string{"echo", "a b", "c d"}},
	{"echo \\$HOME", []string{"echo", "$HOME"}},
	{"echo \\`date\\`", []string{"echo", "`date`"}},
	{"echo \\\\", []string{"echo", "\\"}},
	{"a\\ b", []string{"a\\ b"}},
	{"a\\ b c", []string{"a\\ b", "c"}},
	{"a\\\\ b", []string{"a\\ b"}},
	{"a\\'b", []string{"a'b"}},
	{"echo 'a\nb'", []string{"echo", "a\nb"}},
	{"echo \"a\nb\"", []string{"echo", "a\nb"}},
	{"echo a'b'c", []string{"echo", "abc"}},
	{"echo a\"b\"c", []string{"echo", "abc"}},
	{"café x", []string{"café", "x"}},
	{"你好 世界", []string{"你好", "世界"}},
	{"안녕 하세요", []string{"안녕", "하세요"}},
	{"λογος λογια", []string{"λογος", "λογια"}},
	{"echo #hash", []string{"echo", "#hash"}},
	{"# comment here", []string{"#", "comment", "here"}},
	{"git status | grep modified", []string{"git", "status", "|", "grep", "modified"}},
	{"cat < in > out", []string{"cat", "<", "in", ">", "out"}},
	{"a;b", []string{"a;b"}},
	{"a&&b", []string{"a&&b"}},
	{"a||b", []string{"a||b"}},
	{"cmd --flag=value", []string{"cmd", "--flag=value"}},
	{"cmd -m=value", []string{"cmd", "-m=value"}},
	{"cmd -- -x", []string{"cmd", "--", "-x"}},
	{"echo $VAR", []string{"echo", "$VAR"}},
	{"echo ${VAR}", []string{"echo", "${VAR}"}},
	{"echo ~/file", []string{"echo", "~/file"}},
	{"echo *.go", []string{"echo", "*.go"}},
	{"echo ?", []string{"echo", "?"}},
	{"echo [abc]", []string{"echo", "[abc]"}},
	{"cmd 'a'\"b\"c", []string{"cmd", "abc"}},
	{"cmd \"a'b\"", []string{"cmd", "a'b"}},
	{"cmd 'a \"b'", []string{"cmd", "a \"b"}},
	{"echo \\\"quoted\\\"", []string{"echo", "\"quoted\""}},
	{"echo 'it'\"'\"'s'", []string{"echo", "it's"}},
	{"tar xvf \"foo bar.tar\"", []string{"tar", "xvf", "foo bar.tar"}},
	{"sed -e 's/\\/foo/bar'", []string{"sed", "-e", "s/\\/foo/bar"}},
	{"printf \"a\tb\"", []string{"printf", "a\tb"}},
	{"printf 'a\tb'", []string{"printf", "a\tb"}},
	{"a\t\tb", []string{"a", "b"}},
	{"  \t ", []string{}},
	{"\"a b\" \"c d\"", []string{"a b", "c d"}},
	{"'a b' 'c d'", []string{"a b", "c d"}},
	{"cmd \"\" \"\"", []string{"cmd", "", ""}},
	{"cmd --opt='a b'", []string{"cmd", "--opt=a b"}},
	{"foo=bar baz", []string{"foo=bar", "baz"}},
	{"foo=bar baz=qux", []string{"foo=bar", "baz=qux"}},
	{"a=b c=d", []string{"a=b", "c=d"}},
	{"PATH=/usr/bin:.", []string{"PATH=/usr/bin:."}},
	{"echo ${PATH}", []string{"echo", "${PATH}"}},
	{"git -c user.name=test commit", []string{"git", "-c", "user.name=test", "commit"}},
	{"cmd 'a\\'b'", []string{"cmd", "'a\\'b'"}},
	{"cmd \"a\\\"b\"", []string{"cmd", "a\"b"}},
	{"echo \\n", []string{"echo", "n"}},
	{"echo a\\tb", []string{"echo", "atb"}},
	{"echo 'a\\\\b'", []string{"echo", "a\\\\b"}},
	{"echo \"a\\\\b\"", []string{"echo", "a\\b"}},
	{"echo 'a\\\"b'", []string{"echo", "a\\\"b"}},
	{"cmd \"'\"", []string{"cmd", "'"}},
	{"cmd '\"'", []string{"cmd", "\""}},
	{"x\\\\y", []string{"x\\y"}},
	{"x\\\\\\y", []string{"x\\y"}},
	{"x \\\\ y", []string{"x", "\\ y"}},
	{"x \\' y", []string{"x", "'", "y"}},
	{"a 'unterminated", []string{"a", "'unterminated"}},
	{"a \"unterminated", []string{"a", "\"unterminated"}},
	{"'", []string{"'"}},
	{"\"", []string{"\""}},
	{"echo '", []string{"echo", "'"}},
	{"echo \"", []string{"echo", "\""}},
	{"a 'b c", []string{"a", "'b", "c"}},
	{"a \"b c", []string{"a", "\"b", "c"}},
	{"echo a\\ b", []string{"echo", "a\\ b"}},
	{"echo a\\ b c", []string{"echo", "a\\ b", "c"}},
	{"cd my\\ dir", []string{"cd", "my\\ dir"}},
	{"echo ??", []string{"echo", "\\ "}},
	{"echo a??b", []string{"echo", "a\\ b"}},
	{"ls \\foo bar", []string{"ls", "foo", "bar"}},
	{"echo \\ ", []string{"echo", "\\ "}},
	{"a\\ ", []string{"a\\ "}},
	{"a??b", []string{"a\\ b"}},
	{"??", []string{"\\ "}},
	{"git commit -m 'fix: don't panic'", []string{"git", "commit", "-m", "'fix:", "don't", "panic'"}},
	{"a\\ b\\ c", []string{"a\\ b\\ c"}},
}

// TestGenericSplitCommandPythonFixtures compares Generic.SplitCommand with
// thefuck's Generic.split_command, including its escaped-space workaround.
func TestGenericSplitCommandPythonFixtures(t *testing.T) {
	generic := NewGeneric()
	for index, fixture := range shlexSplitCommandFixtures {
		t.Run(fmt.Sprintf("case_%03d", index), func(t *testing.T) {
			got := generic.SplitCommand(fixture.command)
			if !slices.Equal(got, fixture.want) {
				t.Errorf("SplitCommand(%q) = %q, want %q", fixture.command, got, fixture.want)
			}
		})
	}
}
