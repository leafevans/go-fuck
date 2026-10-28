// Package output obtains the output of a command that failed.
//
// Like the original implementation, go-fuck re-runs the failed command with a
// short timeout to capture its output. The experimental instant mode of the
// reference implementation is not supported yet: there the output is read from
// a terminal log instead of re-running the command.
package output

import (
	"bytes"
	"context"
	"os"
	"strings"
	"time"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/logs"
)

// GetOutput re-runs the expanded command and returns its combined output.
//
// The second return value reports whether the output could be obtained; it is
// false when the command did not finish within the configured timeout.
func GetOutput(script, expanded string) (string, bool) {
	if conf.Current().InstantMode {
		logs.Debug("Instant mode is not supported by go-fuck, re-running the command instead")
	}
	return rerun(expanded)
}

// rerun runs the command and captures its output.
func rerun(expanded string) (string, bool) {
	settings := conf.Current()

	parts := strings.Fields(expanded)
	isSlow := false
	if len(parts) > 0 {
		for _, slow := range settings.SlowCommands {
			if parts[0] == slow {
				isSlow = true
				break
			}
		}
	}

	timeout := settings.WaitCommand
	if isSlow {
		timeout = settings.WaitSlowCommand
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
		err := <-done
		logs.Debug("Command killed after the timeout: %v", err)
		logs.Debug("Execution timed out!")
		return "", false
	}
}

// mergedEnv merges the configured environment over the current environment.
func mergedEnv(overrides map[string]string) []string {
	environment := map[string]string{}
	for _, entry := range os.Environ() {
		if key, value, found := strings.Cut(entry, "="); found {
			environment[key] = value
		}
	}
	for key, value := range overrides {
		environment[key] = value
	}
	merged := make([]string, 0, len(environment))
	for key, value := range environment {
		merged = append(merged, key+"="+value)
	}
	return merged
}
