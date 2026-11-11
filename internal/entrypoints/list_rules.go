package entrypoints

import (
	json "encoding/json/v2"
	"fmt"
	"os"
	"sort"

	"github.com/leafevans/go-fuck/internal/argumentparser"
	"github.com/leafevans/go-fuck/internal/consts"
	"github.com/leafevans/go-fuck/internal/rules"
)

// listRules runs the --list-rules command.
func listRules(args *argumentparser.Args) int {
	names := make([]string, 0, rules.Count())
	for _, rule := range rules.All() {
		names = append(names, rule.Name)
	}
	sort.Strings(names)

	if args.JSON {
		encoded, err := json.Marshal(names)
		if err != nil {
			fmt.Fprintln(os.Stderr, consts.ExecutableName()+": cannot encode the rule list: "+err.Error())
			return 1
		}
		fmt.Fprintln(os.Stdout, string(encoded))
		return 0
	}
	for _, name := range names {
		fmt.Fprintln(os.Stdout, name)
	}
	return 0
}
