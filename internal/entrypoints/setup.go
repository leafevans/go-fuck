package entrypoints

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/leafevans/go-fuck/internal/argumentparser"
	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/system"
)

// setup configures the alias for the current shell. It is the implementation of
// the --setup option.
//
// Shells that can be configured automatically get the required line appended to
// their configuration file. Other shells, such as PowerShell, and shells whose
// configuration file does not exist yet get the line printed instead. Running
// --setup again is safe: an already configured file is not modified.
func setup(args *argumentparser.Args) int {
	if _, err := conf.Load(&conf.Args{Debug: args.Debug}); err != nil {
		fmt.Fprintf(os.Stderr, "go-fuck: can't load settings: %v\n", err)
	}

	shell := shells.Current()
	configuration := shell.HowToConfigure()
	if configuration == nil || configuration.Content == "" || configuration.Path == "" {
		fmt.Fprintf(os.Stderr, "go-fuck can't configure the %s shell automatically.\n", shell.FriendlyName())
		return 1
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
		fmt.Fprintf(os.Stderr, "go-fuck: can't read %s: %v\n", path, err)
		return 1
	}

	if strings.Contains(string(existing), configuration.Content) {
		fmt.Fprintf(os.Stderr, "go-fuck is already configured for %s (%s).\n", shell.FriendlyName(), path)
		fmt.Fprintf(os.Stderr, "Reload it with: %s\n", configuration.Reload)
		return 0
	}

	if err := appendConfigurationLine(path, configuration.Content, existing); err != nil {
		fmt.Fprintf(os.Stderr, "go-fuck: can't update %s: %v\n", path, err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "go-fuck is now configured for %s (%s).\n", shell.FriendlyName(), path)
	fmt.Fprintf(os.Stderr, "Reload it with: %s\n", configuration.Reload)
	return 0
}

// setupConfigPath expands the shell configuration path.
func setupConfigPath(path string) (string, error) {
	expanded := system.ExpandHome(path)
	if !filepath.IsAbs(expanded) {
		return "", fmt.Errorf("not an absolute path: %s", path)
	}
	return expanded, nil
}

// appendConfigurationLine appends content to the existing shell configuration.
func appendConfigurationLine(path, content string, existing []byte) error {
	line := content
	if len(existing) > 0 && existing[len(existing)-1] != '\n' {
		line = "\n" + line
	}
	line += "\n"

	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.WriteString(line); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

// printManualConfiguration prints the line that has to be added manually.
func printManualConfiguration(configuration *shells.Configuration) {
	fmt.Fprintf(os.Stderr, "Add the following line to %s:\n\n  %s\n\n", configuration.Path, configuration.Content)
	fmt.Fprintf(os.Stderr, "Then reload it with: %s\n", configuration.Reload)
}
