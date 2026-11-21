// Package corrector collects the enabled rules and produces the corrected
// commands for a failed command.
package corrector

import (
	"cmp"
	"slices"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/conf"
	"github.com/leafevans/go-fuck/v2/internal/consts"
	"github.com/leafevans/go-fuck/v2/internal/logs"
	"github.com/leafevans/go-fuck/v2/internal/rules"
	"github.com/leafevans/go-fuck/v2/internal/types"
)

// GetRules returns the enabled rules ordered by priority.
//
// Rules listed in the exclude_rules setting are skipped and the priority
// setting overrides a rule priority.
func GetRules() []*types.Rule {
	return configuredRules(conf.Current())
}

func configuredRules(settings *conf.Settings) []*types.Rule {
	enabled := make([]*types.Rule, 0, rules.Count())
	if len(settings.Priority) == 0 {
		walkEnabledRules(settings, func(rule *types.Rule) {
			enabled = append(enabled, rule)
		})
		return enabled
	}
	walkEnabledRules(settings, func(rule *types.Rule) {
		if priority, overridden := settings.Priority[rule.Name]; overridden {
			configured := *rule
			configured.Priority = priority
			rule = &configured
		}
		enabled = append(enabled, rule)
	})
	slices.SortStableFunc(enabled, func(left, right *types.Rule) int {
		return cmp.Compare(left.Priority, right.Priority)
	})
	return enabled
}

func walkEnabledRules(settings *conf.Settings, visit func(*types.Rule)) {
	excluded := make(map[string]struct{}, len(settings.ExcludeRules))
	for _, name := range settings.ExcludeRules {
		excluded[name] = struct{}{}
	}
	defaults := slices.Contains(settings.Rules, consts.AllEnabled)
	for _, rule := range rules.All() {
		if _, skip := excluded[rule.Name]; skip {
			logs.Debug("Ignoring excluded rule: %s", rule.Name)
			continue
		}
		if !(defaults && rule.EnabledByDefault) && !slices.Contains(settings.Rules, rule.Name) {
			continue
		}
		visit(rule)
	}
}

// OrganizeCommands sorts the corrected commands and removes duplicates.
//
// The first command of the input is always returned first, because it is the
// primary suggestion of the highest priority rule.
func OrganizeCommands(commands []*types.CorrectedCommand) []*types.CorrectedCommand {
	return organizeCommands(slices.Clone(commands), conf.Current().Debug)
}

// organizeCommands may reorder and compact its caller-owned slice.
func organizeCommands(ordered []*types.CorrectedCommand, debug bool) []*types.CorrectedCommand {
	if len(ordered) == 0 {
		return nil
	}
	slices.SortStableFunc(ordered[1:], func(left, right *types.CorrectedCommand) int {
		return cmp.Compare(left.Priority, right.Priority)
	})

	organized := ordered[:0]
	const linearDedupLimit = 16
	if len(ordered) <= linearDedupLimit {
		// Most corrections have only a few candidates; skip building a hash index.
		for _, command := range ordered {
			if !slices.ContainsFunc(organized, command.Equal) {
				organized = append(organized, command)
			}
		}
	} else {
		seen := types.NewCorrectedCommandSet(len(ordered))
		for _, command := range ordered {
			if seen.Add(command) {
				organized = append(organized, command)
			}
		}
	}
	clear(ordered[len(organized):])

	if debug {
		descriptions := make([]string, len(organized))
		for i, command := range organized {
			descriptions[i] = command.String()
		}
		logs.Debug("Corrected commands: %s", strings.Join(descriptions, ", "))
	}
	return organized
}

// GetCorrectedCommands returns the sorted and deduplicated corrected commands
// for a failed command.
func GetCorrectedCommands(command *types.Command) []*types.CorrectedCommand {
	settings := conf.Current()
	var corrected []*types.CorrectedCommand
	collect := func(rule *types.Rule) {
		if rule.IsMatch(command) {
			corrected = appendCorrections(corrected, command.Script, rule.Name, rule.GetCorrectedCommands(command))
		}
	}
	if len(settings.Priority) == 0 {
		walkEnabledRules(settings, collect)
	} else {
		for _, rule := range configuredRules(settings) {
			collect(rule)
		}
	}
	return organizeCommands(corrected, settings.Debug)
}

// appendCorrections adds suggestions directly, skipping no-ops without a side effect.
func appendCorrections(dst []*types.CorrectedCommand, original, ruleName string, candidates []*types.CorrectedCommand) []*types.CorrectedCommand {
	for _, candidate := range candidates {
		if candidate.SideEffect == nil && candidate.Script == original {
			continue
		}
		candidate.RuleName = ruleName
		dst = append(dst, candidate)
	}
	return dst
}
