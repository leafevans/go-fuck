package entrypoints

import (
	"errors"

	"github.com/leafevans/go-fuck/internal/instant"
	"github.com/leafevans/go-fuck/internal/logs"
)

// shellLogger runs the experimental instant mode shell logger.
//
// The shell logger replaces the current shell with a logged shell; it only
// returns when that shell exits.
func shellLogger(path string) int {
	code, err := instant.RunShellLogger(path)
	if err != nil {
		if errors.Is(err, instant.ErrUnsupported) {
			logs.Warn("Shell logger doesn't support your platform")
		} else {
			logs.Warn("Shell logger failed: " + err.Error())
		}
		return 1
	}
	return code
}
