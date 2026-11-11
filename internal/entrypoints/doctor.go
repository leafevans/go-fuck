package entrypoints

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/leafevans/go-fuck/internal/argumentparser"
	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/consts"
	"github.com/leafevans/go-fuck/internal/instant"
	"github.com/leafevans/go-fuck/internal/logs"
	"github.com/leafevans/go-fuck/internal/rules"
	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/utils"
)

// doctorStatus is the result of a single doctor check.
type doctorStatus string

const (
	doctorOK   doctorStatus = "ok"
	doctorWarn doctorStatus = "warn"
	doctorFail doctorStatus = "fail"
	doctorInfo doctorStatus = "info"
)

// doctorCheck is one line of the doctor report.
type doctorCheck struct {
	Name    string       `json:"name"`
	Status  doctorStatus `json:"status"`
	Message string       `json:"message"`
	Hint    string       `json:"hint,omitempty"`
}

// doctorReport is the structured output of --doctor.
type doctorReport struct {
	Version   string        `json:"version"`
	GoVersion string        `json:"go_version"`
	Platform  string        `json:"platform"`
	Binary    string        `json:"binary"`
	Shell     string        `json:"shell"`
	Checks    []doctorCheck `json:"checks"`
}

// doctorShellState is the shell information used by the doctor checks.
type doctorShellState struct {
	Name       string
	Info       string
	ConfigPath string
	AliasFound bool
	ReadErr    error
	History    int
}

// The variables below are replaceable in tests.
var (
	doctorExecutable   = os.Executable
	doctorReadFile     = os.ReadFile
	doctorIsAvailable  = utils.IsAvailable
	doctorLoadSettings = func() (*conf.Settings, error) { return conf.Load(&conf.Args{}) }
	doctorCollectShell = collectDoctorShellState
)

// collectDoctorShellState gathers the information about the invoking shell.
func collectDoctorShellState() doctorShellState {
	shell := shells.Current()
	state := doctorShellState{
		Name: shell.Name(),
		Info: shell.Info(),
	}
	if configuration := shell.HowToConfigure(); configuration != nil {
		state.ConfigPath = configuration.Path
		if state.ConfigPath != "" && !strings.HasPrefix(state.ConfigPath, "$") {
			content, err := doctorReadFile(expandHome(state.ConfigPath))
			if err != nil {
				state.ReadErr = err
			} else {
				state.AliasFound = strings.Contains(string(content),
					consts.ExecutableName()+" --alias")
			}
		}
	}
	func() {
		defer func() { _ = recover() }()
		state.History = len(shell.GetHistory())
	}()
	return state
}

// expandHome replaces a leading "~" with the user home directory.
func expandHome(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	return filepath.Join(home, strings.TrimPrefix(path, "~"))
}

// collectDoctorReport runs every doctor check.
func collectDoctorReport() doctorReport {
	report := doctorReport{
		Version:   consts.Version,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}
	if executable, err := doctorExecutable(); err == nil {
		report.Binary = executable
	}
	shell := doctorCollectShell()
	report.Shell = shell.Info
	report.Checks = doctorChecks(report, shell)
	return report
}

// doctorChecks builds the check list of the report.
func doctorChecks(report doctorReport, shell doctorShellState) []doctorCheck {
	checks := make([]doctorCheck, 0, 10)

	checks = append(checks, doctorCheck{
		Name:   "version",
		Status: doctorOK,
		Message: fmt.Sprintf("%s (%s, %s), %d rules", report.Version,
			report.GoVersion, report.Platform, rules.Count()),
	})

	binaryStatus, binaryMessage := doctorOK, report.Binary
	if binaryMessage == "" {
		binaryStatus = doctorFail
		binaryMessage = "could not determine the executable path"
	}
	checks = append(checks, doctorCheck{Name: "binary", Status: binaryStatus, Message: binaryMessage})

	checks = append(checks, doctorShellCheck(shell))
	checks = append(checks, doctorAliasCheck(shell))
	checks = append(checks, doctorSettingsCheck())
	checks = append(checks, doctorHistoryCheck(shell))
	checks = append(checks, doctorFzfCheck())
	checks = append(checks, doctorInstantCheck())
	return checks
}

// doctorShellCheck reports the detected shell.
func doctorShellCheck(shell doctorShellState) doctorCheck {
	if shell.Name == "" || shell.Name == "generic" {
		return doctorCheck{
			Name:    "shell",
			Status:  doctorWarn,
			Message: "could not detect the invoking shell",
			Hint:    "run go-fuck through the shell alias, or set TF_SHELL",
		}
	}
	message := shell.Info
	if name := os.Getenv("TF_SHELL"); name != "" {
		message = fmt.Sprintf("%s (TF_SHELL=%s)", message, name)
	}
	return doctorCheck{Name: "shell", Status: doctorOK, Message: message}
}

// doctorAliasCheck reports whether the alias is configured.
func doctorAliasCheck(shell doctorShellState) doctorCheck {
	check := doctorCheck{Name: "alias"}
	switch {
	case shell.ConfigPath == "":
		check.Status = doctorWarn
		check.Message = "the shell does not have a configuration file to inspect"
		check.Hint = "run go-fuck --alias and add its output to your shell configuration"
	case shell.ConfigPath == "$profile":
		check.Status = doctorInfo
		check.Message = "the PowerShell profile cannot be checked automatically"
		check.Hint = "add iex \"$(" + consts.ExecutableName() + " --alias)\" to $PROFILE"
	case strings.HasPrefix(shell.ConfigPath, "$"):
		check.Status = doctorInfo
		check.Message = "the " + shell.ConfigPath + " profile cannot be checked automatically"
		check.Hint = "the profile must contain: " + consts.ExecutableName() + " --alias"
	case shell.AliasFound:
		check.Status = doctorOK
		check.Message = "found in " + shell.ConfigPath
	default:
		check.Status = doctorWarn
		check.Message = "no alias found in " + shell.ConfigPath
		check.Hint = "run go-fuck --setup to add it"
	}
	return check
}

// doctorSettingsCheck reports whether the settings file can be loaded.
func doctorSettingsCheck() doctorCheck {
	check := doctorCheck{Name: "config"}
	path := filepath.Join(conf.UserDir(), "settings.json")
	if _, err := os.Stat(path); err != nil {
		check.Status = doctorInfo
		check.Message = "no settings file, using the defaults"
		return check
	}
	if _, err := doctorLoadSettings(); err != nil {
		check.Status = doctorFail
		check.Message = "invalid settings file: " + err.Error()
		return check
	}
	check.Status = doctorOK
	check.Message = "loaded " + path
	return check
}

// doctorHistoryCheck reports whether the shell history is readable.
func doctorHistoryCheck(shell doctorShellState) doctorCheck {
	check := doctorCheck{Name: "history"}
	switch {
	case shell.Name == "" || shell.Name == "generic":
		check.Status = doctorInfo
		check.Message = "not checked: the invoking shell is unknown"
	case shell.Name == "powershell":
		check.Status = doctorInfo
		check.Message = "provided by the shell alias (Get-History)"
	case shell.History == 0:
		check.Status = doctorWarn
		check.Message = "no history entries found"
		check.Hint = "go-fuck reads the history of the invoking shell"
	default:
		check.Status = doctorOK
		check.Message = fmt.Sprintf("%d entries", shell.History)
	}
	return check
}

// doctorFzfCheck reports whether the optional fzf selector is available.
func doctorFzfCheck() doctorCheck {
	check := doctorCheck{Name: "fzf"}
	available := doctorIsAvailable("fzf")
	switch {
	case available && conf.Current().Fzf:
		check.Status = doctorOK
		check.Message = "installed and enabled"
	case available:
		check.Status = doctorInfo
		check.Message = "installed, disabled (optional)"
	default:
		check.Status = doctorInfo
		check.Message = "not installed (optional)"
	}
	return check
}

// doctorInstantCheck reports the state of the instant mode and validates the
// terminal log when a logged shell is active.
func doctorInstantCheck() doctorCheck {
	check := doctorCheck{Name: "instant"}
	if !instant.Supported() {
		check.Status = doctorInfo
		check.Message = "not supported on this platform"
		return check
	}

	logPath := instant.LogPath()
	if logPath == "" {
		check.Status = doctorInfo
		check.Message = "available, not enabled (opt-in)"
		check.Hint = "go-fuck --setup --enable-experimental-instant-mode"
		return check
	}

	info, err := os.Stat(logPath)
	if err != nil {
		check.Status = doctorWarn
		check.Message = "enabled, but the terminal log is not readable"
		check.Hint = "restart the shell after running go-fuck --setup"
		return check
	}
	if info.Size() == 0 {
		check.Status = doctorInfo
		check.Message = "enabled (logged shell, no output yet)"
		return check
	}
	if !doctorLogHasMarker(logPath) {
		check.Status = doctorWarn
		check.Message = "enabled, but the terminal log has no prompt marker"
		check.Hint = "restart the shell and keep the instant mode line as the last prompt definition"
		return check
	}
	check.Status = doctorOK
	check.Message = "enabled (logged shell)"
	return check
}

// doctorLogHasMarker reports whether the instant mode log contains the prompt
// marker that the log reader splits command calls at.
func doctorLogHasMarker(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), consts.UserCommandMark)
}

// doctor runs the --doctor command.
func doctor(args *argumentparser.Args) int {
	report := collectDoctorReport()
	if args.JSON {
		encoded, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, consts.ExecutableName()+": cannot encode the report: "+err.Error())
			return 1
		}
		fmt.Fprintln(os.Stdout, string(encoded))
	} else {
		printDoctorReport(report)
	}
	for _, check := range report.Checks {
		if check.Status == doctorFail {
			return 1
		}
	}
	return 0
}

// printDoctorReport prints the report for humans.
func printDoctorReport(report doctorReport) {
	fmt.Fprintln(os.Stdout, logs.Strong("go-fuck doctor"))
	fmt.Fprintln(os.Stdout)
	for _, check := range report.Checks {
		label := fmt.Sprintf("%-6s", string(check.Status))
		switch check.Status {
		case doctorOK:
			label = logs.Success(label)
		case doctorWarn:
			label = logs.Accent(label)
		case doctorFail:
			label = logs.Failure(label)
		default:
			label = logs.Muted(label)
		}
		fmt.Fprintf(os.Stdout, "  %-9s %s %s\n", check.Name, label, check.Message)
		if check.Hint != "" {
			fmt.Fprintf(os.Stdout, "  %-9s %-6s %s\n", "", "", logs.Muted(check.Hint))
		}
	}
}
