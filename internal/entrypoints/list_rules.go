package entrypoints

import (
	json "encoding/json/v2"
	"fmt"
	"os"
	"sort"

	"github.com/leafevans/go-fuck/v2/internal/argumentparser"
	"github.com/leafevans/go-fuck/v2/internal/consts"
	"github.com/leafevans/go-fuck/v2/internal/rules"
)

// listRule is one rule of the machine readable rule list.
type listRule struct {
	// Name is the name accepted by the rules, exclude_rules and priority
	// settings.
	Name string `json:"name"`
	// Priority orders the rule. Lower values are tried first.
	Priority int `json:"priority"`
	// EnabledByDefault reports whether the rule is enabled unless it is listed
	// in the rules setting.
	EnabledByDefault bool `json:"enabled_by_default"`
	// RequiresOutput reports whether the rule needs the command output.
	RequiresOutput bool `json:"requires_output"`
	// SideEffect reports whether the interactive mode runs a side effect before
	// the corrected command.
	SideEffect bool `json:"side_effect"`
}

// listRules runs the --list-rules command.
func listRules(args *argumentparser.Args) int {
	all := rules.All()
	sort.Slice(all, func(left, right int) bool { return all[left].Name < all[right].Name })

	if args.JSON {
		listed := make([]listRule, 0, len(all))
		for _, rule := range all {
			listed = append(listed, listRule{
				Name:             rule.Name,
				Priority:         rule.Priority,
				EnabledByDefault: rule.EnabledByDefault,
				RequiresOutput:   rule.RequiresOutput,
				SideEffect:       rule.SideEffect != nil,
			})
		}
		encoded, err := json.Marshal(listed)
		if err != nil {
			fmt.Fprintln(os.Stderr, consts.ExecutableName()+": cannot encode the rule list: "+err.Error())
			return 1
		}
		fmt.Fprintln(os.Stdout, string(encoded))
		return 0
	}
	for _, rule := range all {
		fmt.Fprintln(os.Stdout, rule.Name)
	}
	return 0
}
