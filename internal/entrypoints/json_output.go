package entrypoints

import (
	"encoding/json"
	"os"

	"github.com/leafevans/go-fuck/internal/logs"
	"github.com/leafevans/go-fuck/internal/types"
)

// jsonSuggestion is one corrected command of the machine readable output.
type jsonSuggestion struct {
	// Command is the corrected command line.
	Command string `json:"command"`
	// Priority orders the suggestions. Lower values are suggested first.
	Priority int `json:"priority"`
	// SideEffect reports whether the normal interactive mode would run a side
	// effect (for example "git add") before the command.
	SideEffect bool `json:"side_effect"`
}

// jsonResult is the machine readable result of a correction request.
type jsonResult struct {
	// Command is the command line that was fixed, with shell aliases expanded.
	Command string `json:"command"`
	// Suggestions are the corrected commands, in the order in which they would
	// be offered.
	Suggestions []jsonSuggestion `json:"suggestions"`
}

// printJSON writes the correction result as JSON.
//
// Unlike the interactive mode it does not run the corrected command, update
// the shell history or apply rule side effects.
func printJSON(command *types.Command, corrected []*types.CorrectedCommand) int {
	result := jsonResult{Suggestions: make([]jsonSuggestion, 0, len(corrected))}
	if command != nil {
		result.Command = command.Script
	}
	for _, suggestion := range corrected {
		result.Suggestions = append(result.Suggestions, jsonSuggestion{
			Command:    suggestion.Script,
			Priority:   suggestion.Priority,
			SideEffect: suggestion.SideEffect != nil,
		})
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		logs.Exception("Can't write the JSON output", err)
		return 1
	}
	return 0
}
