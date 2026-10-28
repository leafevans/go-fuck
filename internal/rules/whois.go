package rules

import (
	"net/url"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// whoisMatch reports whether the command is a whois call with an argument.
func whoisMatch(_ *types.Command) bool {
	// Python: the rule matches every whois command with an argument, the
	// argument is fixed by the correction.
	return true
}

// whoisNewCommand strips the scheme, path and (when possible) the left-most
// subdomain from the whois argument.
func whoisNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	if len(parts) < 2 {
		return nil, nil
	}

	parsed, err := url.Parse(parts[1])
	if err != nil {
		return nil, nil
	}

	if strings.Contains(command.Script, "/") {
		return []string{"whois " + parsed.Host}, nil
	}

	if strings.Contains(command.Script, ".") {
		path := strings.Split(parsed.Path, ".")
		results := make([]string, 0, len(path)-1)
		for index := 1; index < len(path); index++ {
			results = append(results, "whois "+strings.Join(path[index:], "."))
		}
		return results, nil
	}
	// Python returns None for an argument without a path or a dot, which
	// yields no executable suggestion.
	return nil, nil
}

func init() {
	// Python: for_app('whois', at_least=1).
	register(types.NewRule("whois",
		utils.ForApp([]string{"whois"}, 1, whoisMatch),
		whoisNewCommand))
}
