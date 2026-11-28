// Package capture selects and records how the output of a failed command was
// obtained.
package capture

import (
	"time"

	"github.com/leafevans/go-fuck/v4/internal/instant"
	"github.com/leafevans/go-fuck/v4/internal/logs"
	"github.com/leafevans/go-fuck/v4/internal/output"
)

// Mode controls which capture sources may be used.
type Mode string

// Supported capture modes.
const (
	ModeAuto    Mode = "auto"
	ModeInstant Mode = "instant"
	ModeRerun   Mode = "rerun"
)

// Source identifies the source that produced a command's output.
type Source string

// Supported capture sources.
const (
	SourceProvided   Source = "provided"
	SourceInstantLog Source = "instant_log"
	SourceRerun      Source = "rerun"
	SourceNone       Source = "none"
)

// Request describes one command-output capture attempt.
type Request struct {
	// Script is the raw command used to locate the matching instant-log call.
	Script string
	// Expanded is the alias-expanded command used for re-running.
	Expanded string
	// AllowRerun permits the rerun fallback.
	AllowRerun bool
	// Mode selects the capture policy.
	Mode Mode
}

// Result is the output of one capture attempt.
type Result struct {
	Output   string
	Source   Source
	OK       bool
	Duration time.Duration
}

// Capture reads the instant log or reruns the command according to the policy.
func Capture(request Request) (result Result) {
	started := time.Now()
	defer func() {
		result.Duration = time.Since(started)
		logs.Debug("Capture mode=%s source=%s ok=%t took=%s",
			request.Mode, result.Source, result.OK, result.Duration)
	}()

	if request.Mode == ModeAuto || request.Mode == ModeInstant {
		if logPath := instant.LogPath(); logPath != "" {
			if logOutput, ok := instant.LastOutput(logPath, request.Script); ok {
				return Result{Output: logOutput, Source: SourceInstantLog, OK: true}
			}
			logs.Debug("Command not found in the instant mode log")
		} else {
			logs.Debug("No instant mode log path is set")
		}
	}

	if request.Mode != ModeInstant && request.AllowRerun {
		if commandOutput, ok := output.Rerun(request.Expanded); ok {
			return Result{Output: commandOutput, Source: SourceRerun, OK: true}
		}
	}
	return Result{Source: SourceNone}
}
