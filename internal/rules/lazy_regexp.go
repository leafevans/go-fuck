package rules

import (
	"regexp"
	"sync"
)

// lazyRegexp defers the compilation of a rule pattern until the rule is used.
//
// The rules package registers 160 patterns, and compiling them all costs about
// eight milliseconds of every process start, including --help and --list-rules
// where no pattern is needed at all. A pattern is therefore stored as a string
// first and compiled on its first use; the compiled expression is cached, so a
// correction pays the compilation once and the matching itself is unchanged.
type lazyRegexp struct {
	pattern  string
	once     sync.Once
	compiled *regexp.Regexp
}

// newRegexp returns a lazily compiled regular expression.
func newRegexp(pattern string) *lazyRegexp {
	return &lazyRegexp{pattern: pattern}
}

// regexp returns the compiled expression, compiling it on the first call.
func (r *lazyRegexp) expression() *regexp.Regexp {
	r.once.Do(func() {
		r.compiled = regexp.MustCompile(r.pattern)
	})
	return r.compiled
}

// The methods below mirror the part of *regexp.Regexp that the rules use.
func (r *lazyRegexp) MatchString(s string) bool { return r.expression().MatchString(s) }

func (r *lazyRegexp) FindString(s string) string { return r.expression().FindString(s) }

func (r *lazyRegexp) FindStringSubmatch(s string) []string {
	return r.expression().FindStringSubmatch(s)
}

func (r *lazyRegexp) FindStringSubmatchIndex(s string) []int {
	return r.expression().FindStringSubmatchIndex(s)
}

func (r *lazyRegexp) FindAllStringSubmatch(s string, n int) [][]string {
	return r.expression().FindAllStringSubmatch(s, n)
}

func (r *lazyRegexp) ReplaceAllString(src, replacement string) string {
	return r.expression().ReplaceAllString(src, replacement)
}

func (r *lazyRegexp) Split(s string, n int) []string { return r.expression().Split(s, n) }

func (r *lazyRegexp) Match(b []byte) bool { return r.expression().Match(b) }

func (r *lazyRegexp) SubexpIndex(name string) int { return r.expression().SubexpIndex(name) }

func (r *lazyRegexp) String() string { return r.pattern }
