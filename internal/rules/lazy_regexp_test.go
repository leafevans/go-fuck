package rules

import (
	"slices"
	"sync"
	"testing"
)

func TestLazyRegexpDelegatesToCompiledExpression(t *testing.T) {
	expression := newRegexp(`(?P<name>[a-z]+)-([0-9]+)`)
	if got, want := expression.String(), `(?P<name>[a-z]+)-([0-9]+)`; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
	if !expression.MatchString("abc-12") {
		t.Fatal("MatchString() did not match")
	}
	if got, want := expression.FindString("before abc-12 after"), "abc-12"; got != want {
		t.Fatalf("FindString() = %q, want %q", got, want)
	}
	if got, want := expression.FindStringSubmatch("abc-12"), []string{"abc-12", "abc", "12"}; !slices.Equal(got, want) {
		t.Fatalf("FindStringSubmatch() = %q, want %q", got, want)
	}
	if got, want := expression.FindStringSubmatchIndex("abc-12"), []int{0, 6, 0, 3, 4, 6}; !slices.Equal(got, want) {
		t.Fatalf("FindStringSubmatchIndex() = %v, want %v", got, want)
	}
	if got := expression.FindAllStringSubmatch("abc-1 xyz-22", -1); len(got) != 2 || got[1][0] != "xyz-22" {
		t.Fatalf("FindAllStringSubmatch() = %q", got)
	}
	if got, want := expression.ReplaceAllString("abc-12", "$1"), "abc"; got != want {
		t.Fatalf("ReplaceAllString() = %q, want %q", got, want)
	}
	if expression.Match([]byte("not matching")) {
		t.Fatal("Match() matched unrelated bytes")
	}
	if !expression.Match([]byte("abc-3")) {
		t.Fatal("Match() did not match bytes")
	}
	if got, want := expression.SubexpIndex("name"), 1; got != want {
		t.Fatalf("SubexpIndex(name) = %d, want %d", got, want)
	}
	if got := expression.SubexpIndex("missing"); got != -1 {
		t.Fatalf("SubexpIndex(missing) = %d, want -1", got)
	}
	if got, want := newRegexp(",").Split("a,b,c", -1), []string{"a", "b", "c"}; !slices.Equal(got, want) {
		t.Fatalf("Split() = %q, want %q", got, want)
	}
	first := expression.expression()
	second := expression.expression()
	if first != second {
		t.Fatal("expression() recompiled the pattern")
	}
}

func TestLazyRegexpIsSafeForConcurrentUse(t *testing.T) {
	expression := newRegexp(`^git\s+push`)
	var workers sync.WaitGroup
	for range 16 {
		workers.Go(func() {
			if !expression.MatchString("git push origin main") {
				t.Error("concurrent MatchString() did not match")
			}
		})
	}
	workers.Wait()
}
