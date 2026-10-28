package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// cpp11Match reports whether the compiler complained about a missing C++11
// flag.
func cpp11Match(command *types.Command) bool {
	return strings.Contains(command.Output,
		"This file requires compiler and library support for the ISO C++ 2011 standard.") ||
		strings.Contains(command.Output, "-Wc++11-extensions")
}

// cpp11NewCommand enables the C++11 standard.
func cpp11NewCommand(command *types.Command) ([]string, error) {
	return []string{command.Script + " -std=c++11"}, nil
}

func init() {
	register(types.NewRule("cpp11",
		utils.ForApp([]string{"g++", "clang++"}, 0, cpp11Match),
		cpp11NewCommand))
}
