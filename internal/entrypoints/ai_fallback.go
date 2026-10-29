package entrypoints

import (
	"github.com/leafevans/go-fuck/internal/ai"
	"github.com/leafevans/go-fuck/internal/consts"
	"github.com/leafevans/go-fuck/internal/logs"
	"github.com/leafevans/go-fuck/internal/types"
)

// aiFallback asks the optional AI endpoint for a correction when no rule
// matched the command. It returns nil when the fallback is disabled or fails.
func aiFallback(command *types.Command) []*types.CorrectedCommand {
	client := ai.FromSettings()
	if client == nil {
		return nil
	}

	logs.Debug("No rule matched, asking the AI fallback")
	suggestion, err := client.Suggest(command)
	if err != nil {
		logs.Warn("AI fallback failed: " + err.Error())
		return nil
	}
	if suggestion == "" {
		return nil
	}

	return []*types.CorrectedCommand{{
		Script:   suggestion,
		Priority: consts.DefaultPriority,
	}}
}
