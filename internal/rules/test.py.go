package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
)

// testDotPyMatch reports whether a script named test.py could not be run.
func testDotPyMatch(command *types.Command) bool {
	return command.Script == "test.py" && strings.Contains(command.Output, "not found")
}

// testDotPyNewCommand runs the tests with pytest.
func testDotPyNewCommand(_ *types.Command) ([]string, error) {
	return []string{"pytest"}, nil
}

func init() {
	rule := types.NewRule("test.py", testDotPyMatch, testDotPyNewCommand)
	// Python: priority = 900, so that it comes before the python_command rule.
	rule.Priority = 900
	register(rule)
}
