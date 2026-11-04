package entrypoints

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/leafevans/go-fuck/internal/argumentparser"
	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/logs"
	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/system"
)

// powerShellProfileExecutables are the PowerShell flavors whose profile files
// are configured, in display order.
var powerShellProfileExecutables = []string{"powershell.exe", "pwsh"}

// powerShellProfiles returns the profile files of the installed PowerShell
// flavors. It is a variable so tests can replace the subprocess lookup.
var powerShellProfiles = func() []string {
	return resolveProfiles(powerShellProfileExecutables, powerShellProfile)
}

// setup configures the alias for the current shell. It is the implementation of
// the --setup option.
//
// Shells that can be configured automatically get the required line appended to
// their configuration file. Other shells and shells whose configuration file
// does not exist yet get the line printed instead. Running --setup again is
// safe: an already configured file is not modified.
//
// PowerShell has one profile file per installed flavor, so the installed
// executables are asked for their $PROFILE path and every profile that can be
// written to is updated.
//
// With --enable-experimental-instant-mode the instant mode initialization line
// is configured instead of the regular alias. Shells without instant mode
// support fall back to the regular alias and report a warning.
func setup(args *argumentparser.Args) int {
	if _, err := conf.Load(&conf.Args{Debug: args.Debug}); err != nil {
		fmt.Fprintln(os.Stderr, logs.FailureLine(fmt.Sprintf("can't load settings: %v", err)))
	}

	shell := shells.Current()
	configuration := shell.HowToConfigure()
	if configuration == nil || configuration.Content == "" || configuration.Path == "" {
		fmt.Fprintln(os.Stderr, logs.WarningLine(fmt.Sprintf("can't configure the %s shell automatically", shell.FriendlyName())))
		return 1
	}

	if args.EnableExperimentalInstantMode {
		if configuration.InstantContent == "" {
			fmt.Fprintln(os.Stderr, logs.WarningLine(fmt.Sprintf("instant mode is not supported for %s, configuring the regular alias instead", shell.FriendlyName())))
		} else {
			instantConfiguration := *configuration
			instantConfiguration.Content = configuration.InstantContent
			configuration = &instantConfiguration
		}
	}

	if shell.Name() == "powershell" {
		return setupPowerShell(shell, configuration)
	}

	if !configuration.CanConfigureAutomatically || strings.HasPrefix(configuration.Path, "$") {
		printManualConfiguration(configuration)
		return 0
	}

	path, err := setupConfigPath(configuration.Path)
	if err != nil {
		printManualConfiguration(configuration)
		return 0
	}

	existing, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			printManualConfiguration(configuration)
			return 0
		}
		fmt.Fprintln(os.Stderr, logs.FailureLine(fmt.Sprintf("can't read %s: %v", path, err)))
		return 1
	}

	if strings.Contains(string(existing), configuration.Content) {
		printConfigured(shell.FriendlyName(), path, true)
		fmt.Fprintf(os.Stderr, "%s %s\n", logs.Muted("Reload it with:"), logs.Strong(configuration.Reload))
		return 0
	}

	// Older releases wrote the instant mode line without quoting the eval
	// argument. Replace that line in place so enabling instant mode again does
	// not append a second initialization line.
	if args.EnableExperimentalInstantMode {
		if changed, err := replaceInstantModeLine(path, existing, configuration.Content); err != nil {
			fmt.Fprintln(os.Stderr, logs.FailureLine(fmt.Sprintf("can't update %s: %v", path, err)))
			return 1
		} else if changed {
			printConfigured(shell.FriendlyName(), path, false)
			fmt.Fprintf(os.Stderr, "%s %s\n", logs.Muted("Reload it with:"), logs.Strong(configuration.Reload))
			return 0
		}
	}

	if err := appendConfigurationLine(path, configuration.Content, existing); err != nil {
		fmt.Fprintln(os.Stderr, logs.FailureLine(fmt.Sprintf("can't update %s: %v", path, err)))
		return 1
	}

	printConfigured(shell.FriendlyName(), path, false)
	fmt.Fprintf(os.Stderr, "%s %s\n", logs.Muted("Reload it with:"), logs.Strong(configuration.Reload))
	return 0
}

// setupPowerShell configures the alias in the profile file of every installed
// PowerShell flavor.
//
// The profile path is not a constant: Windows PowerShell 5.1 and PowerShell 7
// use different directories and the Documents folder may be redirected (for
// example to OneDrive), so the installed executables report their $PROFILE
// instead of go-fuck guessing it. Missing profile files and their directories
// are created, because a fresh PowerShell installation does not have them.
func setupPowerShell(shell shells.Shell, configuration *shells.Configuration) int {
	paths := powerShellProfiles()
	if len(paths) == 0 {
		printManualConfiguration(configuration)
		return 0
	}

	configured := 0
	alreadyConfigured := 0
	for _, path := range paths {
		existing, err := os.ReadFile(path)
		if err != nil && !os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, logs.FailureLine(fmt.Sprintf("can't read %s: %v", path, err)))
			return 1
		}
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				fmt.Fprintln(os.Stderr, logs.FailureLine(fmt.Sprintf("can't create %s: %v", filepath.Dir(path), err)))
				return 1
			}
			existing = nil
		}
		if strings.Contains(string(existing), configuration.Content) {
			printConfigured(shell.FriendlyName(), path, true)
			alreadyConfigured++
			continue
		}
		if err := appendConfigurationLine(path, configuration.Content, existing); err != nil {
			fmt.Fprintln(os.Stderr, logs.FailureLine(fmt.Sprintf("can't update %s: %v", path, err)))
			return 1
		}
		printConfigured(shell.FriendlyName(), path, false)
		configured++
	}

	if configured > 0 || alreadyConfigured > 0 {
		fmt.Fprintf(os.Stderr, "%s %s\n", logs.Muted("Reload it with:"), logs.Strong(configuration.Reload))
	}
	return 0
}

// resolveProfiles resolves configuration paths with lookup, keeps the first
// occurrence of every path and drops flavors that are not installed.
func resolveProfiles(executables []string, lookup func(string) (string, bool)) []string {
	seen := make(map[string]bool, len(executables))
	paths := make([]string, 0, len(executables))
	for _, executable := range executables {
		path, ok := lookup(executable)
		if !ok || seen[path] {
			continue
		}
		seen[path] = true
		paths = append(paths, path)
	}
	return paths
}

// powerShellProfile returns the $PROFILE path reported by the PowerShell
// executable, or false when the executable is not installed or does not
// report a profile path.
func powerShellProfile(executable string) (string, bool) {
	command, err := exec.LookPath(executable)
	if err != nil {
		command = wellKnownPowerShellPath(executable)
		if command == "" {
			return "", false
		}
	}
	profile, err := exec.Command(command, "-NoProfile", "-NonInteractive",
		"-Command", "[Console]::OutputEncoding=[System.Text.Encoding]::UTF8; $PROFILE").Output()
	if err != nil {
		return "", false
	}
	path := strings.TrimSpace(string(profile))
	if path == "" {
		return "", false
	}
	return path, true
}

// wellKnownPowerShellPath returns the default installation path of Windows
// PowerShell. The directory is usually on PATH, but some environments run
// go-fuck with a reduced PATH, so the well-known location is used as fallback.
func wellKnownPowerShellPath(executable string) string {
	if executable != "powershell.exe" {
		return ""
	}
	root := os.Getenv("SystemRoot")
	if root == "" {
		return ""
	}
	path := filepath.Join(root, "System32", "WindowsPowerShell", "v1.0", executable)
	if _, err := os.Stat(path); err != nil {
		return ""
	}
	return path
}

// setupConfigPath expands the shell configuration path.
func setupConfigPath(path string) (string, error) {
	expanded := system.ExpandHome(path)
	if !filepath.IsAbs(expanded) {
		return "", fmt.Errorf("not an absolute path: %s", path)
	}
	return expanded, nil
}

// replaceInstantModeLine replaces a legacy instant mode initialization line with
// content and reports whether the file was changed.
func replaceInstantModeLine(path string, existing []byte, content string) (bool, error) {
	lines := strings.Split(string(existing), "\n")
	changed := false
	for index, line := range lines {
		if line != content && strings.Contains(line, "--enable-experimental-instant-mode") {
			lines[index] = content
			changed = true
		}
	}
	if !changed {
		return false, nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return true, os.WriteFile(path, []byte(strings.Join(lines, "\n")), info.Mode().Perm())
}

// appendConfigurationLine appends content to the existing shell configuration.
func appendConfigurationLine(path, content string, existing []byte) error {
	line := content
	if len(existing) > 0 && existing[len(existing)-1] != '\n' {
		line = "\n" + line
	}
	line += "\n"

	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.WriteString(line); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

// printConfigured reports that the shell configuration was found or updated.
func printConfigured(shellName, path string, alreadyConfigured bool) {
	state := "is now configured for"
	if alreadyConfigured {
		state = "is already configured for"
	}
	fmt.Fprintf(os.Stderr, "%s %s %s (%s).\n",
		logs.Strong("go-fuck"), state, logs.Accent(shellName), logs.Muted(path))
}

// printManualConfiguration prints the line that has to be added manually.
func printManualConfiguration(configuration *shells.Configuration) {
	fmt.Fprintf(os.Stderr, "%s %s\n\n  %s\n\n%s %s\n",
		logs.Strong("Add the following line to"), logs.Muted(configuration.Path),
		logs.Strong(configuration.Content),
		logs.Muted("Then reload it with:"), logs.Strong(configuration.Reload))
}
