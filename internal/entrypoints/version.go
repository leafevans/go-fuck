package entrypoints

import (
	"runtime"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/consts"
	"github.com/leafevans/go-fuck/internal/instant"
	"github.com/leafevans/go-fuck/internal/logs"
	"github.com/leafevans/go-fuck/internal/rules"
	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/utils"
)

// versionInfo collects the details printed by --version.
//
// Detecting the shell version can cost a process launch: PowerShell, for
// example, is queried by running "pwsh --version". The plain one line printed
// to a pipe or a script only needs the shell name, so the version is detected
// only when detailed is true.
func versionInfo(detailed bool) logs.VersionInfo {
	instantState := "not supported on this platform"
	if instant.Supported() {
		instantState = "available (bash/zsh on Linux and macOS)"
		if instant.LogPath() != "" {
			instantState = "enabled (logged shell)"
		}
	}
	shell := shells.Current().FriendlyName()
	if detailed {
		shell = shells.Current().Info()
	}
	return logs.VersionInfo{
		Version:     consts.Version,
		GoVersion:   runtime.Version(),
		Shell:       shell,
		Platform:    runtime.GOOS + "/" + runtime.GOARCH,
		Rules:       rules.Count(),
		ConfigPath:  conf.UserDir(),
		InstantMode: instantState,
		Fzf:         utils.Which("fzf") != "",
		Alias:       conf.Alias(),
	}
}
