// Package corrector collects the enabled rules and produces the corrected
// commands for a failed command.
package corrector

import (
	"sort"
	"strings"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/logs"
	"github.com/leafevans/go-fuck/internal/rules"
	"github.com/leafevans/go-fuck/internal/types"
)

// GetRules returns the enabled rules ordered by priority.
//
// Rules listed in the exclude_rules setting are skipped and the priority
// setting overrides a rule priority.
func GetRules() []*types.Rule {
	settings := conf.Current()
	excluded := map[string]bool{}
	for _, name := range settings.ExcludeRules {
		excluded[name] = true
	}

	enabled := []*types.Rule{}
	for _, rule := range rules.All() {
		if excluded[rule.Name] {
			logs.Debug("Ignoring excluded rule: %s", rule.Name)
			continue
		}
		if priority, ok := settings.Priority[rule.Name]; ok {
			rule.Priority = priority
		}
		if rule.IsEnabled() {
			enabled = append(enabled, rule)
		}
	}

	sort.SliceStable(enabled, func(left, right int) bool {
		return enabled[left].Priority < enabled[right].Priority
	})
	return enabled
}

// OrganizeCommands sorts the corrected commands and removes duplicates.
//
// The first command of the input is always returned first, because it is the
// primary suggestion of the highest priority rule.
func OrganizeCommands(commands []*types.CorrectedCommand) []*types.CorrectedCommand {
	if len(commands) == 0 {
		return nil
	}

	first := commands[0]
	rest := make([]*types.CorrectedCommand, len(commands)-1)
	copy(rest, commands[1:])
	sort.SliceStable(rest, func(left, right int) bool {
		return rest[left].Priority < rest[right].Priority
	})

	seen := []*types.CorrectedCommand{first}
	organized := []*types.CorrectedCommand{first}
	for _, command := range rest {
		duplicate := false
		for _, existing := range seen {
			if command.Equal(existing) {
				duplicate = true
				break
			}
		}
		if duplicate {
			continue
		}
		seen = append(seen, command)
		organized = append(organized, command)
	}

	descriptions := make([]string, len(organized))
	for i, command := range organized {
		descriptions[i] = command.String()
	}
	logs.Debug("Corrected commands: %s", strings.Join(descriptions, ", "))
	return organized
}

// GetCorrectedCommands returns the sorted and deduplicated corrected commands
// for a failed command.
func GetCorrectedCommands(command *types.Command) []*types.CorrectedCommand {
	corrected := []*types.CorrectedCommand{}
	for _, rule := range GetRules() {
		if !rule.IsMatch(command) {
			continue
		}
		corrected = append(corrected, rule.GetCorrectedCommands(command)...)
	}
	return OrganizeCommands(corrected)
}
