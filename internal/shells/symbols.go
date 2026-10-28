package shells

// Characters and separators used to build shell snippets without escape
// sequences in string literals.
var (
	// newline is the line separator.
	newline = string(rune(10))
	// backtick is used by the tcsh alias.
	backtick = string(rune(96))
	// dquote is used by templates that embed shell variables in double quotes.
	dquote = string(rune(34))
	// backslash is used to escape characters for nested shells.
	backslash = string(rune(92))
)
