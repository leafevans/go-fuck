package entrypoints

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/leafevans/go-fuck/v4/internal/argumentparser"
	"github.com/leafevans/go-fuck/v4/internal/conf"
	"github.com/leafevans/go-fuck/v4/internal/consts"
	"github.com/leafevans/go-fuck/v4/internal/logs"
	"github.com/leafevans/go-fuck/v4/internal/shells"
	"github.com/leafevans/go-fuck/v4/internal/system"
)

// errAmbiguousInitialization reports that more than one known alias line would
// be changed, so setup cannot migrate the file safely.
var errAmbiguousInitialization = errors.New("multiple go-fuck initialization lines found")

// errPowerShellInitializationConflict reports that PowerShell profiles need
// manual cleanup before setup can make a safe change.
var errPowerShellInitializationConflict = errors.New("PowerShell initialization conflict")

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
	renderer := logs.NewRenderer(os.Stdout)
	fmt.Fprintf(os.Stdout, "%s\n\n", renderer.Title("setup"))

	if _, err := conf.Load(&conf.Args{Debug: args.Debug, CaptureMode: args.CaptureMode}); err != nil {
		printSetupFailure("can't load settings: %v", err)
	}

	shell := shells.Current()
	configuration := shell.HowToConfigure()
	if configuration == nil || configuration.Content == "" || configuration.Path == "" {
		printSetupWarning("can't configure the %s shell automatically", shell.FriendlyName())
		return 1
	}

	mode := conf.Current().CaptureMode
	if args.CaptureMode == "" && !args.EnableExperimentalInstantMode &&
		!conf.Current().CaptureModeExplicit() {
		mode = "auto"
	}
	if args.EnableExperimentalInstantMode {
		mode = "instant"
	}
	rawRegularContent := configuration.Content
	rawInstantContent := configuration.InstantContent
	regularContent := withCaptureMode(rawRegularContent, mode)
	instantContent := withCaptureMode(rawInstantContent, mode)
	regularConfiguration := *configuration
	regularConfiguration.Content = regularContent
	configuration = &regularConfiguration
	if (mode == "auto" || mode == "instant") && configuration.InstantContent != "" {
		instantConfiguration := regularConfiguration
		instantConfiguration.Content = instantContent
		configuration = &instantConfiguration
	} else if mode == "instant" {
		printSetupWarning("instant capture is not supported for %s, configuring rerun mode instead", shell.FriendlyName())
		mode = "rerun"
		regularContent = withCaptureMode(rawRegularContent, mode)
		regularConfiguration.Content = regularContent
		configuration = &regularConfiguration
	}

	if shell.Name() == "powershell" {
		return setupPowerShell(shell, configuration, regularContent, instantContent,
			rawRegularContent, rawInstantContent)
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
		printSetupFailure("can't read %s: %v", path, err)
		return 1
	}

	if hasConfigurationLine(existing, configuration.Content) {
		printConfigured(shell.FriendlyName(), path, true)
		printReload(configuration.Reload)
		return 0
	}

	if changed, err := replaceInitializationLine(path, existing, configuration.Content,
		regularContent, configuration.InstantContent, rawRegularContent, rawInstantContent); err != nil {
		printSetupFailure("can't update %s: %v", path, err)
		return 1
	} else if changed {
		printConfigured(shell.FriendlyName(), path, false)
		printReload(configuration.Reload)
		return 0
	}

	// Older releases wrote the instant mode line without quoting the eval
	// argument. Replace that line in place so enabling instant mode again does
	// not append a second initialization line.
	if args.EnableExperimentalInstantMode {
		if changed, err := replaceInstantModeLine(path, existing, configuration.Content); err != nil {
			printSetupFailure("can't update %s: %v", path, err)
			return 1
		} else if changed {
			printConfigured(shell.FriendlyName(), path, false)
			printReload(configuration.Reload)
			return 0
		}
	}

	if err := appendConfigurationLine(path, configuration.Content, existing); err != nil {
		printSetupFailure("can't update %s: %v", path, err)
		return 1
	}

	printConfigured(shell.FriendlyName(), path, false)
	printReload(configuration.Reload)
	return 0
}

// withCaptureMode records the selected capture policy in an initialization
// line so later shell starts do not fall back to an older settings file.
func withCaptureMode(content, mode string) string {
	if content == "" || mode == "" {
		return content
	}
	return strings.Replace(content, "--alias", "--alias --capture-mode="+mode, 1)
}

// setupPowerShell configures the alias in the profile file of every installed
// PowerShell flavor.
//
// The profile path is not a constant: Windows PowerShell 5.1 and PowerShell 7
// use different directories and the Documents folder may be redirected (for
// example to OneDrive), so the installed executables report their $PROFILE
// instead of go-fuck guessing it. Missing profile files and their directories
// are created, because a fresh PowerShell installation does not have them.
func setupPowerShell(shell shells.Shell, configuration *shells.Configuration,
	regularContent, instantContent, rawRegularContent, rawInstantContent string) int {
	paths := powerShellProfiles()
	if len(paths) == 0 {
		printManualConfiguration(configuration)
		return 0
	}

	pairs, err := inspectPowerShellProfiles(paths, regularContent, instantContent,
		rawRegularContent, rawInstantContent)
	if err != nil {
		if errors.Is(err, errPowerShellInitializationConflict) {
			printPowerShellInitializationConflict(configuration)
		}
		return 1
	}

	configured := 0
	alreadyConfigured := 0
	for _, pair := range pairs {
		if pair.initialized == nil {
			if err := os.MkdirAll(filepath.Dir(pair.host), 0o755); err != nil {
				printSetupFailure("can't create %s: %v", filepath.Dir(pair.host), err)
				return 1
			}
			if err := appendConfigurationLine(pair.host, configuration.Content, nil); err != nil {
				printSetupFailure("can't update %s: %v", pair.host, err)
				return 1
			}
			printConfigured(shell.FriendlyName(), pair.host, false)
			configured++
			continue
		}
		if hasConfigurationLine(pair.initialized.existing, configuration.Content) {
			printConfigured(shell.FriendlyName(), pair.initialized.path, true)
			alreadyConfigured++
			continue
		}
		if changed, err := replaceInitializationLine(pair.initialized.path, pair.initialized.existing,
			configuration.Content,
			regularContent, instantContent, rawRegularContent, rawInstantContent); err != nil {
			printSetupFailure("can't update %s: %v", pair.initialized.path, err)
			return 1
		} else if changed {
			printConfigured(shell.FriendlyName(), pair.initialized.path, false)
			configured++
			continue
		}
	}

	if configured > 0 || alreadyConfigured > 0 {
		printReload(configuration.Reload)
	}
	return 0
}

// powerShellProfilePair describes the two current-user profiles PowerShell
// loads for one host: the all-hosts profile and the current-host profile.
type powerShellProfilePair struct {
	host        string
	initialized *powerShellProfileState
}

// powerShellProfileState is one profile file that contains an initialization.
type powerShellProfileState struct {
	path     string
	existing []byte
}

// inspectPowerShellProfiles checks every all-hosts/current-host pair before
// setup writes anything. It returns a conflict error for multiple recognized
// initializations or a custom initialization that cannot be migrated safely.
func inspectPowerShellProfiles(hostProfiles []string, candidates ...string) ([]powerShellProfilePair, error) {
	pairs := make([]powerShellProfilePair, 0, len(hostProfiles))
	conflict := false
	seen := make(map[string]bool)
	for _, host := range hostProfiles {
		if seen[host] {
			continue
		}
		seen[host] = true

		pair := powerShellProfilePair{host: host}
		initialized := []powerShellProfileState{}
		multiple := []string{}
		custom := []string{}
		paths := []string{filepath.Join(filepath.Dir(host), "profile.ps1"), host}
		for index, path := range paths {
			if index == 1 && path == paths[0] {
				continue
			}
			existing, err := os.ReadFile(path)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				printSetupFailure("can't read %s: %v", path, err)
				return nil, err
			}
			switch lineCount := countInitializationLines(existing, candidates...); {
			case lineCount == 1:
				initialized = append(initialized, powerShellProfileState{path: path, existing: existing})
			case lineCount > 1:
				multiple = append(multiple, path)
			}
			if hasCustomInitialization(existing, candidates...) {
				custom = append(custom, path)
			}
		}

		switch {
		case len(multiple) > 0:
			printSetupWarning("multiple go-fuck initializations found in %s; keep exactly one",
				strings.Join(multiple, ", "))
			conflict = true
		case len(initialized) > 1:
			printSetupWarning("multiple go-fuck initializations found in %s; keep exactly one",
				strings.Join(profilePaths(initialized), ", "))
			conflict = true
		case len(custom) > 0:
			printSetupWarning("custom go-fuck initialization found in %s; replace it with the direct initialization line",
				strings.Join(custom, ", "))
			conflict = true
		case len(initialized) == 1:
			pair.initialized = &initialized[0]
		}
		pairs = append(pairs, pair)
	}
	if conflict {
		return nil, errPowerShellInitializationConflict
	}
	return pairs, nil
}

// profilePaths returns the paths of initialized profile states.
func profilePaths(states []powerShellProfileState) []string {
	paths := make([]string, len(states))
	for index, state := range states {
		paths[index] = state.path
	}
	return paths
}

// countInitializationLines counts known go-fuck initialization lines. A line
// matching the current and a legacy spelling is counted once.
func countInitializationLines(existing []byte, candidates ...string) int {
	allCandidates := initializationCandidates(candidates...)
	count := 0
	for line := range strings.SplitSeq(string(existing), "\n") {
		for _, candidate := range allCandidates {
			if _, ok := configurationLineSuffix(line, candidate); ok {
				count++
				break
			}
		}
	}
	return count
}

// hasCustomInitialization reports whether a profile references go-fuck's alias
// through code that setup does not know how to migrate safely. Comments and
// recognized direct initialization lines are ignored.
func hasCustomInitialization(existing []byte, candidates ...string) bool {
	allCandidates := initializationCandidates(candidates...)
	needle := consts.ExecutableName() + " --alias"
	for line := range strings.SplitSeq(string(existing), "\n") {
		trimmed := strings.TrimLeft(line, " \t\uFEFF")
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		known := false
		for _, candidate := range allCandidates {
			if _, ok := configurationLineSuffix(line, candidate); ok {
				known = true
				break
			}
		}
		if !known && strings.Contains(line, needle) {
			return true
		}
	}
	return false
}

// initializationCandidates expands every known initialization spelling.
func initializationCandidates(candidates ...string) []string {
	allCandidates := make([]string, 0, len(candidates)*2)
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		allCandidates = append(allCandidates, candidate)
		if quoted, ok := strings.CutPrefix(candidate, `eval "`); ok && strings.HasSuffix(quoted, `"`) {
			allCandidates = append(allCandidates, "eval "+strings.TrimSuffix(quoted, `"`))
		}
	}
	return allCandidates
}

// printPowerShellInitializationConflict explains how to replace a custom or
// duplicate initialization with one direct line.
func printPowerShellInitializationConflict(configuration *shells.Configuration) {
	printSetupWarning("setup made no changes; keep one initialization in your PowerShell profiles and reload it")
	fmt.Fprintf(os.Stderr, "  %s\n", configuration.Content)
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

// hasConfigurationLine reports whether content appears as an initialization
// line, optionally indented or followed by a comment. Mentions in comments or
// other commands do not configure a shell.
func hasConfigurationLine(existing []byte, content string) bool {
	for line := range strings.SplitSeq(string(existing), "\n") {
		if _, ok := configurationLineSuffix(line, content); ok {
			return true
		}
	}
	return false
}

// configurationLineSuffix matches a complete initialization command and keeps
// any trailing whitespace or comment for callers that replace the command.
func configurationLineSuffix(line, content string) (string, bool) {
	line = strings.TrimLeft(line, " \t\uFEFF")
	suffix, ok := strings.CutPrefix(line, content)
	if !ok {
		return "", false
	}
	rest := strings.TrimSpace(suffix)
	return suffix, rest == "" || strings.HasPrefix(rest, "#")
}

// replaceInstantModeLine upgrades the legacy unquoted eval initialization.
// Unrelated commands, comments, indentation and line endings are preserved.
func replaceInstantModeLine(path string, existing []byte, content string) (bool, error) {
	quoted, ok := strings.CutPrefix(content, `eval "`)
	legacy := ""
	if ok && strings.HasSuffix(quoted, `"`) {
		legacy = "eval " + strings.TrimSuffix(quoted, `"`)
	}
	return replaceInitializationLine(path, existing, content, legacy)
}

// replaceInitializationLine replaces exactly one known initialization line
// while preserving indentation and trailing comments.
func replaceInitializationLine(path string, existing []byte, content string, candidates ...string) (bool, error) {
	allCandidates := initializationCandidates(candidates...)
	lines := strings.Split(string(existing), "\n")
	matches := 0
	for index, line := range lines {
		for _, candidate := range allCandidates {
			if candidate == "" || candidate == content {
				continue
			}
			suffix, matched := configurationLineSuffix(line, candidate)
			if !matched {
				continue
			}
			indent := line[:len(line)-len(strings.TrimLeft(line, " \t\uFEFF"))]
			lines[index] = indent + content + suffix
			matches++
			break
		}
	}
	if matches == 0 {
		return false, nil
	}
	if matches > 1 {
		return false, errAmbiguousInitialization
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
	renderer := logs.NewRenderer(os.Stdout)
	state := "is now configured for"
	if alreadyConfigured {
		state = "is already configured for"
	}
	fmt.Fprintf(os.Stdout, "%s %s %s %s\n",
		renderer.StatusLabel(logs.StatusOK), renderer.Strong(shellName), state, renderer.Muted(path))
}

// printManualConfiguration prints the line that has to be added manually.
func printManualConfiguration(configuration *shells.Configuration) {
	renderer := logs.NewRenderer(os.Stdout)
	fmt.Fprintf(os.Stdout, "%s %s %s\n\n  %s\n\n%s %s\n",
		renderer.StatusLabel(logs.StatusInfo), "Add the following line to", renderer.Muted(configuration.Path),
		renderer.Strong(configuration.Content),
		renderer.Muted("Then reload with:"), renderer.Strong(configuration.Reload))
}

// printReload prints the command that reloads the shell configuration.
func printReload(command string) {
	renderer := logs.NewRenderer(os.Stdout)
	fmt.Fprintf(os.Stdout, "%s %s\n", renderer.Muted("Reload with:"), renderer.Strong(command))
}

// printSetupFailure reports a failed setup step on stderr.
func printSetupFailure(format string, args ...any) {
	renderer := logs.NewRenderer(os.Stderr)
	fmt.Fprintln(os.Stderr, renderer.Diagnostic(logs.StatusFail, fmt.Sprintf(format, args...)))
}

// printSetupWarning reports a recoverable setup warning on stderr.
func printSetupWarning(format string, args ...any) {
	renderer := logs.NewRenderer(os.Stderr)
	fmt.Fprintln(os.Stderr, renderer.Diagnostic(logs.StatusWarn, fmt.Sprintf(format, args...)))
}
