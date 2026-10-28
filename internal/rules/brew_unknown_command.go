package rules

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// brewUnknownCommandFallbackCommands is used when the local Homebrew
// installation cannot be inspected. It mirrors the built-in list of the
// original implementation (based on Homebrew 0.9.5).
var brewUnknownCommandFallbackCommands = []string{
	"info", "home", "options", "install", "uninstall",
	"search", "list", "update", "upgrade", "pin", "unpin",
	"doctor", "create", "edit", "cask",
}

// brewUnknownCommandRegexp matches the command that brew did not recognize.
var brewUnknownCommandRegexp = regexp.MustCompile(`Error: Unknown command: ([a-z]+)`)

// brewUnknownCommandBrewCommands returns the commands supported by the local
// Homebrew installation, including the commands provided by taps.
var brewUnknownCommandBrewCommands = utils.Memoize(func() []string {
	prefix := specific.GetBrewPathPrefix()
	if prefix != "" {
		if commands, err := brewUnknownCommandList(prefix); err == nil {
			return commands
		}
	}
	return append([]string{}, brewUnknownCommandFallbackCommands...)
})

// brewUnknownCommandList returns the default and the tap specific brew
// commands.
func brewUnknownCommandList(prefix string) ([]string, error) {
	commands, err := brewUnknownCommandDefaultCommands(prefix)
	if err != nil {
		return nil, err
	}
	tapCommands, err := brewUnknownCommandTapCommands(prefix)
	if err != nil {
		return nil, err
	}
	return append(commands, tapCommands...), nil
}

// brewUnknownCommandDefaultCommands returns the built-in brew commands.
func brewUnknownCommandDefaultCommands(prefix string) ([]string, error) {
	commandPath := filepath.Join(prefix, "Homebrew", "Library", "Homebrew", "cmd")
	entries, err := os.ReadDir(commandPath)
	if err != nil {
		return nil, err
	}
	commands := []string{}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(name, ".rb") || strings.HasSuffix(name, ".sh") {
			commands = append(commands, name[:len(name)-3])
		}
	}
	return commands, nil
}

// brewUnknownCommandTapCommands returns the commands provided by the installed
// taps.
func brewUnknownCommandTapCommands(prefix string) ([]string, error) {
	tapsPath := filepath.Join(prefix, "Homebrew", "Library", "Taps")
	users, err := brewUnknownCommandDirectoryNames(tapsPath)
	if err != nil {
		return nil, err
	}
	commands := []string{}
	for _, user := range users {
		taps, err := brewUnknownCommandDirectoryNames(filepath.Join(tapsPath, user))
		if err != nil {
			return nil, err
		}
		for _, tap := range taps {
			// Tap commands are only provided by taps that follow the
			// "homebrew-" naming convention.
			if !strings.HasPrefix(tap, "homebrew-") {
				continue
			}
			commandPath := filepath.Join(tapsPath, user, tap, "cmd")
			info, err := os.Stat(commandPath)
			if err != nil || !info.IsDir() {
				continue
			}
			entries, err := os.ReadDir(commandPath)
			if err != nil {
				continue
			}
			for _, entry := range entries {
				name := entry.Name()
				if strings.HasPrefix(name, "brew-") && strings.HasSuffix(name, ".rb") {
					name = strings.ReplaceAll(name, "brew-", "")
					name = strings.ReplaceAll(name, ".rb", "")
					commands = append(commands, name)
				}
			}
		}
	}
	return commands, nil
}

// brewUnknownCommandDirectoryNames returns the names of the subdirectories of
// a directory.
func brewUnknownCommandDirectoryNames(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	names := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names, nil
}

// brewUnknownCommandMatch reports whether brew did not recognize the command.
func brewUnknownCommandMatch(command *types.Command) bool {
	if !strings.Contains(command.Script, "brew") ||
		!strings.Contains(command.Output, "Unknown command") {
		return false
	}
	match := brewUnknownCommandRegexp.FindStringSubmatch(command.Output)
	if match == nil {
		return false
	}
	return utils.GetClosestDefault(match[1], brewUnknownCommandBrewCommands()) != ""
}

// brewUnknownCommandNewCommand suggests the closest valid brew commands.
func brewUnknownCommandNewCommand(command *types.Command) ([]string, error) {
	match := brewUnknownCommandRegexp.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	return utils.ReplaceCommand(command, match[1], brewUnknownCommandBrewCommands()), nil
}

func init() {
	rule := types.NewRule("brew_unknown_command", brewUnknownCommandMatch, brewUnknownCommandNewCommand)
	rule.EnabledByDefault = specific.BrewAvailable()
	register(rule)
}
