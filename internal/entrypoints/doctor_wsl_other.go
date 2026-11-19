//go:build !windows

package entrypoints

// doctorWSLInteropCheck is only meaningful for Windows builds.
func doctorWSLInteropCheck() (doctorCheck, bool) {
	return doctorCheck{}, false
}
