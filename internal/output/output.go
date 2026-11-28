// Package output re-runs failed commands with a timeout to capture their output.
// The capture package decides whether a command may be re-run.
package output

import (
	"bytes"
	"context"
	"os"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/leafevans/go-fuck/v4/internal/conf"
	"github.com/leafevans/go-fuck/v4/internal/logs"
)

// Rerun runs the command and captures its output.
func Rerun(expanded string) (string, bool) {
	settings := conf.Current()

	timeout := settings.WaitCommand
	for name := range strings.FieldsSeq(expanded) {
		if slices.Contains(settings.SlowCommands, name) {
			timeout = settings.WaitSlowCommand
		}
		break // Only the first field is the command name.
	}

	stopTimer := logs.DebugTime("Call: " + expanded)
	defer stopTimer()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := shellCommand(expanded)
	cmd.Env = mergedEnv(settings.Env)
	var buffer bytes.Buffer
	cmd.Stdout = &buffer
	cmd.Stderr = &buffer

	if err := cmd.Start(); err != nil {
		logs.Debug("Unable to start the command: %v", err)
		return "", false
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			logs.Debug("Command finished with an error: %v", err)
		}
		result := buffer.String()
		logs.Debug("Received output: %s", result)
		return result, true
	case <-ctx.Done():
		killTree(cmd)
		select {
		case err := <-done:
			logs.Debug("Command killed after the timeout: %v", err)
		case <-time.After(2 * time.Second):
			// A process that escaped the killed process tree can keep the
			// output pipe open; do not block go-fuck on it.
			logs.Debug("Command did not exit after the timeout, ignoring its output")
		}
		logs.Debug("Execution timed out!")
		return "", false
	}
}

// mergedEnv overlays settings on the inherited environment, keeping the
// spelling of existing names even when Windows matches them case-insensitively.
func mergedEnv(overrides map[string]string) []string {
	environment := os.Environ()
	if len(overrides) == 0 {
		return environment
	}
	indices := make(map[string]int, len(environment))
	for i, entry := range environment {
		if key, _, ok := strings.Cut(entry, "="); ok && key != "" {
			indices[envKey(key)] = i
		}
	}
	for key, value := range overrides {
		normalized := envKey(key)
		if i, ok := indices[normalized]; ok {
			name, _, _ := strings.Cut(environment[i], "=")
			environment[i] = name + "=" + value
		} else {
			indices[normalized] = len(environment)
			environment = append(environment, key+"="+value)
		}
	}
	return environment
}

// envKey normalizes an environment variable name for the platform. Windows
// environment variable names are case insensitive.
func envKey(key string) string {
	if runtime.GOOS == "windows" {
		return strings.ToLower(key)
	}
	return key
}
