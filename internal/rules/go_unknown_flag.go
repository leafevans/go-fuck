package rules

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

var (
	// goUndefinedFlag matches the error line of the Go toolchain for an
	// unknown flag, for example: flag provided but not defined: -raec.
	goUndefinedFlag = regexp.MustCompile(`flag provided but not defined: (-[A-Za-z0-9-]+)`)
	// goHelpFlag matches a flag line in the output of "go help <command>".
	goHelpFlag = regexp.MustCompile(`(?m)^\t(-[A-Za-z][A-Za-z0-9-]*)`)
)

// goHelpOutput returns the output of "go help <subcommand>". It is a
// variable so that tests can stub the Go toolchain.
var goHelpOutput = func(subcommand string) string {
	output, err := exec.Command("go", "help", subcommand).Output()
	if err != nil {
		return ""
	}
	return string(output)
}

// goUnknownFlagMatch matches the error that the Go toolchain prints for an
// unknown command line flag.
func goUnknownFlagMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "flag provided but not defined: -")
}

// goUnknownFlagNewCommand replaces the unknown flag with the closest flag
// that the help output of the subcommand lists.
func goUnknownFlagNewCommand(command *types.Command) ([]string, error) {
	match := goUndefinedFlag.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	token := match[1]

	subcommand := ""
	parts := command.ScriptParts()
	if len(parts) > 1 && !strings.HasPrefix(parts[1], "-") {
		subcommand = parts[1]
	}
	if subcommand == "" {
		subcommand = "build"
	}

	flags := []string{}
	for _, flagMatch := range goHelpFlag.FindAllStringSubmatch(goHelpOutput(subcommand), -1) {
		flags = append(flags, flagMatch[1])
	}
	// A cutoff above the default keeps low confidence suggestions such as
	// -relase -> -race out; only clearly similar flags are suggested.
	closest := utils.GetClosest(token, flags, 0.7, false)
	if closest == "" {
		return nil, nil
	}

	broken := ""
	for _, part := range parts {
		if part == token || part == "-"+token {
			broken = part
			break
		}
	}
	if broken == "" {
		return nil, nil
	}
	return []string{utils.ReplaceArgument(command.Script, broken, closest)}, nil
}

func init() {
	register(types.NewRule(
		"go_unknown_flag",
		utils.ForApp([]string{"go"}, 0, goUnknownFlagMatch),
		goUnknownFlagNewCommand,
	))
}
