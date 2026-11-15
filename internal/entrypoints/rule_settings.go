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
