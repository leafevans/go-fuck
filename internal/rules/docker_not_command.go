package rules

import (
	"bytes"
	"os/exec"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// dockerNotCommandWrong matches the misspelled command reported by docker.
var dockerNotCommandWrong = regexp.MustCompile(`docker: '(\w+)' is not a docker command\.`)

// dockerNotCommandMatch reports whether docker did not recognize a command.
func dockerNotCommandMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "is not a docker command") ||
		strings.Contains(command.Output, "Usage:\tdocker")
}

// dockerNotCommandParseCommands extracts the command names of a help section.
func dockerNotCommandParseCommands(lines []string, startsWith string) []string {
	index := 0
	for index < len(lines) && !strings.HasPrefix(lines[index], startsWith) {
		index++
	}
	if index == len(lines) {
		return nil
	}

	commands := []string{}
	for _, line := range lines[index+1:] {
		// A blank line ends the section.
		if strings.TrimSpace(line) == "" {
			break
		}
		commands = append(commands, strings.Split(strings.TrimSpace(line), " ")[0])
	}
	return commands
}

// dockerNotCommandSplitLines splits captured output into lines.
func dockerNotCommandSplitLines(output string) []string {
	if output == "" {
		return nil
	}
	return strings.Split(output, newline)
}

// dockerNotCommandHelpCommands parses the output of "docker".
func dockerNotCommandHelpCommands(stdout, stderr string) []string {
	lines := dockerNotCommandSplitLines(stdout)
	// Old docker versions write their help to stdout, newer ones to stderr.
	if len(lines) == 0 {
		lines = dockerNotCommandSplitLines(stderr)
	}

	managementCommands := []string{}
	if utils.Contains(lines, "Management Commands:") {
		managementCommands = dockerNotCommandParseCommands(lines, "Management Commands:")
	}
	return append(managementCommands, dockerNotCommandParseCommands(lines, "Commands:")...)
}

// dockerNotCommandReadHelp runs docker and captures its help output.
var dockerNotCommandReadHelp = func() (string, string) {
	var stdout, stderr bytes.Buffer
	docker := exec.Command("docker")
	docker.Stdout = &stdout
	docker.Stderr = &stderr
	_ = docker.Run()
	return stdout.String(), stderr.String()
}

// dockerNotCommandCommands returns the commands listed by "docker --help".
var dockerNotCommandCommands = utils.Memoize(func() []string {
	stdout, stderr := dockerNotCommandReadHelp()
	return dockerNotCommandHelpCommands(stdout, stderr)
})

// dockerNotCommandNewCommand suggests the closest valid docker command.
func dockerNotCommandNewCommand(command *types.Command) ([]string, error) {
	if strings.Contains(command.Output, "Usage:") {
		parts := command.ScriptParts()
		// Python indexes script_parts[2] here. Commands without a
		// management subcommand raise IndexError there, so they produce no
		// suggestion either way.
		if len(parts) <= 2 {
			return nil, nil
		}
		subcommands := dockerNotCommandParseCommands(strings.Split(command.Output, newline), "Commands:")
		return utils.ReplaceCommand(command, parts[2], subcommands), nil
	}

	matches := dockerNotCommandWrong.FindAllStringSubmatch(command.Output, -1)
	if len(matches) == 0 {
		return nil, nil
	}
	return utils.ReplaceCommand(command, matches[0][1], dockerNotCommandCommands()), nil
}

func init() {
	register(types.NewRule("docker_not_command",
		specific.SudoSupportMatch(utils.ForApp([]string{"docker"}, 0, dockerNotCommandMatch)),
		specific.SudoSupportNew(dockerNotCommandNewCommand)))
}
