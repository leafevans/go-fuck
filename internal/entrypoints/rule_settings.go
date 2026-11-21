package entrypoints

import (
	"sort"
	"strconv"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/conf"
	"github.com/leafevans/go-fuck/v2/internal/consts"
	"github.com/leafevans/go-fuck/v2/internal/logs"
	"github.com/leafevans/go-fuck/v2/internal/rules"
)

// unknownRuleSettings returns the settings entries that name an unknown rule.
//
// A typo silently changes the behavior: an excluded rule keeps running, a rule
// listed in rules is never enabled, and a priority override is never applied.
// The names are validated against the rule registry so the mistake is reported
// instead of being ignored.
func unknownRuleSettings(settings *conf.Settings) []string {
	if settings == nil {
		return nil
	}
	unknown := []string{}
	seen := map[string]bool{}
	add := func(setting, name string) {
		name = strings.TrimSpace(name)
		if name == "" || name == consts.AllEnabled {
			return
		}
		key := setting + "\x00" + name
		if seen[key] || rules.Get(name) != nil {
			return
		}
		seen[key] = true
		unknown = append(unknown, setting+" "+strconv.Quote(name))
	}
	for _, name := range settings.Rules {
		add("rules", name)
	}
	for _, name := range settings.ExcludeRules {
		add("exclude_rules", name)
	}
	for name := range settings.Priority {
		add("priority", name)
	}
	sort.Strings(unknown)
	return unknown
}

// warnUnknownRuleSettings reports rule names of the settings that do not
// exist.
func warnUnknownRuleSettings(settings *conf.Settings) {
	unknown := unknownRuleSettings(settings)
	if len(unknown) == 0 {
		return
	}
	logs.Warn("Unknown rule names in settings.json (ignored): " + strings.Join(unknown, ", "))
}

// conflictingRuleSettings returns rule settings entries that contradict each
// other.
//
// exclude_rules wins over rules and priority, so a rule listed in both places
// is either redundant (rules has no effect) or ignored (the priority override
// never applies). The contradictions are reported because the settings look
// right while behaving differently.
func conflictingRuleSettings(settings *conf.Settings) []string {
	if settings == nil {
		return nil
	}
	excluded := map[string]bool{}
	for _, name := range settings.ExcludeRules {
		name = strings.TrimSpace(name)
		if name == "" || rules.Get(name) == nil {
			continue
		}
		excluded[name] = true
	}
	listed := map[string]bool{}
	allEnabled := false
	for _, name := range settings.Rules {
		name = strings.TrimSpace(name)
		switch {
		case name == consts.AllEnabled:
			allEnabled = true
		case name == "" || rules.Get(name) == nil:
		default:
			listed[name] = true
		}
	}
	conflicts := []string{}
	seen := map[string]bool{}
	add := func(message string) {
		if seen[message] {
			return
		}
		seen[message] = true
		conflicts = append(conflicts, message)
	}
	for _, name := range settings.Rules {
		name = strings.TrimSpace(name)
		if name == "" || name == consts.AllEnabled || !excluded[name] {
			continue
		}
		add("rules " + strconv.Quote(name) + " also appears in exclude_rules")
	}
	priorities := make([]string, 0, len(settings.Priority))
	for name := range settings.Priority {
		priorities = append(priorities, name)
	}
	sort.Strings(priorities)
	for _, name := range priorities {
		name = strings.TrimSpace(name)
		rule := rules.Get(name)
		if rule == nil {
			continue
		}
		switch {
		case excluded[name]:
			add("priority " + strconv.Quote(name) + " is ignored because the rule is excluded")
		case !listed[name] && (!allEnabled || !rule.DefaultEnabled()):
			add("priority " + strconv.Quote(name) + " is ignored because the rule is not enabled")
		}
	}
	sort.Strings(conflicts)
	return conflicts
}

// warnConflictingRuleSettings reports rule settings that contradict each
// other.
func warnConflictingRuleSettings(settings *conf.Settings) {
	conflicts := conflictingRuleSettings(settings)
	if len(conflicts) == 0 {
		return
	}
	logs.Warn("Conflicting rule settings in settings.json: " + strings.Join(conflicts, ", "))
}
