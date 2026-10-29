package entrypoints

import (
	"path/filepath"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/customrules"
	"github.com/leafevans/go-fuck/internal/logs"
	"github.com/leafevans/go-fuck/internal/rules"
)

// loadCustomRules loads the declarative rules from the "rules" directory of
// the go-fuck configuration directory. Invalid files are reported and skipped
// so that a broken custom rule never prevents the compiled in rules from
// running.
func loadCustomRules() {
	dir := filepath.Join(conf.UserDir(), "rules")
	loaded, problems := customrules.Load(dir)
	for _, problem := range problems {
		logs.Warn("Custom rule: " + problem.Error())
	}
	for _, rule := range loaded {
		if err := rules.RegisterCustom(rule); err != nil {
			logs.Warn("Custom rule: " + err.Error())
		}
	}
	if len(loaded) > 0 {
		logs.Debug("Loaded %d custom rule(s) from %s", len(loaded), dir)
	}
}
