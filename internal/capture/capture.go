// Package capture selects and records how the output of a failed command was
// obtained.
package capture

import (
	"time"

	"github.com/leafevans/go-fuck/v3/internal/instant"
	"github.com/leafevans/go-fuck/v3/internal/logs"
	"github.com/leafevans/go-fuck/v3/internal/output"
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
	// Provided is non-nil when the caller supplied the output. A non-nil empty
	// string is a valid captured output that printed nothing.
	Provided *string
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

// Capture applies the configured source policy.
func Capture(request Request) Result {
	started := time.Now()
	result := Result{Source: SourceNone}
	defer func() {
		result.Duration = time.Since(started)
		logs.Debug("Capture mode=%s source=%s ok=%t took=%s",
			request.Mode, result.Source, result.OK, result.Duration)
	}()

	if request.Provided != nil {
		result.Output = *request.Provided
		result.Source = SourceProvided
		result.OK = true
		return result
	}

	if request.Mode == ModeAuto || request.Mode == ModeInstant {
		if logPath := instant.LogPath(); logPath != "" {
			if logOutput, ok := instant.LastOutput(logPath, request.Script); ok {
				result.Output = logOutput
				result.Source = SourceInstantLog
				result.OK = true
				return result
			}
			logs.Debug("Command not found in the instant mode log")
		} else {
			logs.Debug("No instant mode log path is set")
		}
	}

	if request.Mode == ModeInstant || !request.AllowRerun {
		return result
	}

	if commandOutput, ok := output.Rerun(request.Expanded); ok {
		result.Output = commandOutput
		result.Source = SourceRerun
		result.OK = true
	}
	return result
}
