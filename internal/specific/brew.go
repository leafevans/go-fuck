package specific

import (
	"os/exec"
	"strings"

	"github.com/leafevans/go-fuck/internal/utils"
)

// GetBrewPathPrefix returns the Homebrew installation prefix, or an empty
// string when it cannot be detected.
//
// The result is memoized because detecting the prefix starts a process.
var GetBrewPathPrefix = utils.Memoize(func() string {
	output, err := exec.Command("brew", "--prefix").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
})
