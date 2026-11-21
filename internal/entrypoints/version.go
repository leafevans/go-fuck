package entrypoints

import (
	"os"
	"runtime"

	"github.com/leafevans/go-fuck/v2/internal/conf"
	"github.com/leafevans/go-fuck/v2/internal/consts"
	"github.com/leafevans/go-fuck/v2/internal/instant"
	"github.com/leafevans/go-fuck/v2/internal/logs"
	"github.com/leafevans/go-fuck/v2/internal/rules"
	"github.com/leafevans/go-fuck/v2/internal/shells"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

func printVersion(verbose bool) {
	if !logs.OutputIsTerminal() {
		shell := shells.Current()
		name := shell.FriendlyName()
		if verbose {
			name = shell.Info()
		}
		logs.Version(consts.Version, runtime.Version(), name)
		return
	}
	logs.PrintVersion(versionInfo(verbose), verbose)
}

// versionInfo collects only the fields displayed by the terminal banner.
func versionInfo(verbose bool) logs.VersionInfo {
	info := logs.VersionInfo{
		Version:   consts.Version,
		GoVersion: runtime.Version(),
		Shell:     shells.Current().Info(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
		Rules:     rules.Count(),
	}
	if !verbose {
		return info
	}
	instantState := "not supported on this platform"
	if instant.Supported() {
		instantState = "available (bash/zsh on Linux and macOS)"
		if runtime.GOOS == "windows" {
			instantState = "available (PowerShell on Windows)"
		}
		if instant.LogPath() != "" {
			instantState = "enabled (logged shell)"
		}
	}
	if executable, err := os.Executable(); err == nil {
		info.Executable = executable
	}
	modulePath, moduleVersion := consts.BuildModule()
	info.Module = modulePath
	if moduleVersion != "" && moduleVersion != "(devel)" {
		info.Module += " " + moduleVersion
	}
	info.ConfigPath = conf.UserDir()
	info.InstantMode = instantState
	info.Fzf = utils.IsAvailable("fzf")
	info.Alias = conf.Alias()
	return info
}
