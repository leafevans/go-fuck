package entrypoints

import (
	json "encoding/json/v2"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/argumentparser"
	"github.com/leafevans/go-fuck/v2/internal/consts"
	"github.com/leafevans/go-fuck/v2/internal/logs"
	"github.com/leafevans/go-fuck/v2/internal/rules"
	"github.com/leafevans/go-fuck/v2/internal/types"
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
	all := slices.Clone(rules.All())
	slices.SortFunc(all, func(left, right *types.Rule) int {
		return strings.Compare(left.Name, right.Name)
	})

	if args.JSON {
		listed := make([]listRule, 0, len(all))
		for _, rule := range all {
			listed = append(listed, listRule{
				Name:             rule.Name,
				Priority:         rule.Priority,
				EnabledByDefault: rule.DefaultEnabled(),
				RequiresOutput:   rule.RequiresOutput,
				SideEffect:       rule.SideEffect != nil,
			})
		}
		encoded, err := json.Marshal(listed)
		if err != nil {
			renderer := logs.NewRenderer(os.Stderr)
			fmt.Fprintln(os.Stderr, renderer.Diagnostic(logs.StatusFail,
				consts.ExecutableName()+": cannot encode the rule list: "+err.Error()))
			return 1
		}
		fmt.Fprintln(os.Stdout, string(encoded))
		return 0
	}
	printTextRules(os.Stdout, all, logs.IsTerminal(os.Stdout))
	return 0
}

// printTextRules prints the sorted rule names. A terminal gets a short report
// heading; redirected output remains one rule name per line.
func printTextRules(w io.Writer, all []*types.Rule, terminal bool) {
	renderer := logs.NewRenderer(w)
	if terminal {
		fmt.Fprintln(w, renderer.Title("rules"))
		fmt.Fprintln(w)
		fmt.Fprintln(w, renderer.Diagnostic(logs.StatusInfo, fmt.Sprintf("%d built-in rules", len(all))))
		fmt.Fprintln(w)
	}
	for _, rule := range all {
		if terminal {
			fmt.Fprint(w, "  ")
		}
		fmt.Fprintln(w, rule.Name)
	}
}
