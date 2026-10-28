package rules

import "testing"

// cpp11StandardError is the error printed by old compilers.
const cpp11StandardError = "This file requires compiler and library support for the ISO C++ 2011 standard."

func TestRuleCpp11(t *testing.T) {
	resetTestState(t)
	rule := Get("cpp11")

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "g++ foo.cc", cpp11StandardError)
		assertMatch(t, rule, "clang++ foo.cc", "warning: '-Wc++11-extensions' is enabled")
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "g++ foo.cc", "")
		assertNotMatch(t, rule, "gcc foo.cc", cpp11StandardError)
	})

	t.Run("get new command", func(t *testing.T) {
		assertNewCommands(t, rule, "g++ foo.cc", cpp11StandardError,
			[]string{"g++ foo.cc -std=c++11"})
		assertNewCommands(t, rule, "clang++ foo.cc", "-Wc++11-extensions",
			[]string{"clang++ foo.cc -std=c++11"})
	})
}
