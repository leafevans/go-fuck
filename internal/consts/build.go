package consts

import "runtime/debug"

// BuildModule returns the module path and the module version recorded in the
// running binary.
//
// The module path identifies the release series: this source tree reports
// "github.com/leafevans/go-fuck/v2", while binaries installed from the pre-v2
// module path "github.com/leafevans/go-fuck" belong to the old 0.x and 1.x
// series. The version is "(devel)" for local builds and for release binaries
// that were built outside a module version.
func BuildModule() (path, version string) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", ""
	}
	return info.Main.Path, info.Main.Version
}
