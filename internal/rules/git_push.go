package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v3/internal/shells"
	"github.com/leafevans/go-fuck/v3/internal/specific"
	"github.com/leafevans/go-fuck/v3/internal/types"
)

// gitPushUpstream matches the commands suggested by git.
var gitPushUpstream = newRegexp(`(?m)^[ \t]*git push --set-upstream ([^\s-][^\s]*) ([^\s-][^\s]*)[ \t]*\r?$`)

// gitPushMatch reports whether git suggested setting an upstream branch.
func gitPushMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "git push --set-upstream") && gitHasSubcommand(command.Script, "push")
}

// gitPushNewCommand uses the upstream command suggested by git.
func gitPushNewCommand(command *types.Command) ([]string, error) {
	hints := gitPushUpstream.FindAllStringSubmatch(command.Output, 2)
	if len(hints) != 1 {
		return nil, nil
	}
	shell := shells.Current()
	var script strings.Builder
	end, operands := 0, 0
	for span, word := range specific.GitCommandWords(command.Script, shell) {
		if end == 0 {
			if word != "push" {
				return nil, nil
			}
			script.WriteString(command.Script[:span[1]])
			script.WriteString(" --set-upstream ")
			script.WriteString(shell.Quote(hints[0][1]))
			script.WriteByte(' ')
			script.WriteString(shell.Quote(hints[0][2]))
		} else {
			switch word {
			case "-u", "--set-upstream":
			case "--quiet", "-q", "--verbose", "-v", "--force", "-f", "--force-with-lease", "--dry-run", "-n", "--porcelain", "--atomic", "--no-verify":
				script.WriteString(command.Script[end:span[1]])
			default:
				// A separate option value or shell operator cannot be safely
				// distinguished from a refspec in an arbitrary command.
				if strings.HasPrefix(word, "-") || strings.ContainsAny(word, ";&|<>()`$") || operands == 2 {
					return nil, nil
				}
				operands++
			}
		}
		end = span[1]
	}
	if end == 0 || strings.TrimSpace(command.Script[end:]) != "" {
		return nil, nil
	}
	return []string{script.String() + command.Script[end:]}, nil
}

func init() {
	register(types.NewRule("git_push",
		specific.GitSupportMatch(gitPushMatch),
		specific.GitSupportNew(gitPushNewCommand)))
}
