//go:build windows

package entrypoints

import (
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/system"
)

// doctorAncestorNames returns the names of the process ancestors. It is a
// variable so that tests can replace the lookup.
var doctorAncestorNames = system.AncestorNames

// doctorWSLInteropCheck reports a Windows build that was started from inside
// WSL.
//
// Such a binary reports PowerShell, because its ancestors are the Windows host
// processes of the interop layer, even though the user is in a Linux shell. The
// Linux build is the right one there, so a warning with a hint is more useful
// than the detected shell alone.
func doctorWSLInteropCheck() (doctorCheck, bool) {
	for _, name := range doctorAncestorNames() {
		switch strings.ToLower(name) {
		case "wslhost.exe", "wslrelay.exe", "wsl.exe":
			return doctorCheck{
				Name:    "wsl",
				Status:  doctorWarn,
				Message: "this is the Windows build, started from inside WSL",
				Hint:    "the shell detection sees the Windows host, so use the Linux build from the distribution (for example ~/.local/bin/go-fuck)",
			}, true
		}
	}
	return doctorCheck{}, false
}
