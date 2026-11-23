package ui

import (
	"time"

	"github.com/leafevans/go-fuck/v3/internal/capture"
	"github.com/leafevans/go-fuck/v3/internal/logs"
	"github.com/leafevans/go-fuck/v3/internal/types"
)

// ExplainCommands renders the capture and execution plan without applying it.
func ExplainCommands(command *types.Command, commands []*types.CorrectedCommand) {
	script := ""
	source := capture.SourceNone
	var elapsed time.Duration
	if command != nil {
		script = command.Script
		source = command.CaptureSource()
		elapsed = command.CaptureDuration()
	}
	logs.ExplainHeader(script, string(source), elapsed)
	if len(commands) == 0 {
		logs.ExplainEmpty()
		return
	}
	for index, correction := range commands {
		logs.ExplainSuggestion(index+1, correction.Script, correction.RuleName,
			correction.Explanation(), correction.BeforeRun())
	}
}
