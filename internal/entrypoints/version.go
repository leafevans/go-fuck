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
func versionInfo() logs.VersionInfo {
	instantState := "not supported on this platform"
	if instant.Supported() {
		instantState = "available (bash/zsh on Linux and macOS)"
		if instant.LogPath() != "" {
			instantState = "enabled (logged shell)"
		}
	}
	return logs.VersionInfo{
		Version:     consts.Version,
		GoVersion:   runtime.Version(),
		Shell:       shells.Current().Info(),
		Platform:    runtime.GOOS + "/" + runtime.GOARCH,
		Rules:       rules.Count(),
		ConfigPath:  conf.UserDir(),
		InstantMode: instantState,
		Fzf:         utils.Which("fzf") != "",
		Alias:       conf.Alias(),
	}
}
