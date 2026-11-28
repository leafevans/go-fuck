package specific

import "slices"

// ToolAvailability checks the names used by rules for each tool.
type ToolAvailability func(string) bool

// Probe returns a lazy check for any of the given executable names.
func (available ToolAvailability) Probe(programs ...string) func() bool {
	return func() bool {
		return slices.ContainsFunc(programs, available)
	}
}
