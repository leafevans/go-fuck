// Package conf loads and exposes go-fuck settings.
//
// Settings come from three sources, in increasing order of precedence:
//
//  1. built-in defaults;
//  2. the user configuration file (settings.json in the go-fuck config dir);
//  3. environment variables;
//  4. command line arguments.
//
// Environment variables use the GOFUCK_ prefix. The legacy THEFUCK_ prefix of
// the original implementation is still recognized for compatibility.
package conf

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/leafevans/go-fuck/internal/consts"
)

// Settings is the complete go-fuck configuration.
//
// The zero value is not usable; call Default to obtain a populated value.
type Settings struct {
	// Rules lists the rules that are enabled. The special value
	// consts.AllEnabled enables every rule.
	Rules []string `json:"rules"`
	// ExcludeRules lists rule names that must never be loaded.
	ExcludeRules []string `json:"exclude_rules"`
	// WaitCommand is the number of seconds to wait for a re-run command.
	WaitCommand int `json:"wait_command"`
	// RequireConfirmation asks the user before executing a corrected command.
	RequireConfirmation bool `json:"require_confirmation"`
	// NoColors disables colored output.
	NoColors bool `json:"no_colors"`
	// Debug enables debug output on stderr.
	Debug bool `json:"debug"`
	// Priority overrides rule priorities. Lower values are tried first.
	Priority map[string]int `json:"priority"`
	// HistoryLimit limits the number of shell history lines that are read.
	// Zero means "no limit".
	HistoryLimit int `json:"history_limit"`
	// AlterHistory writes the corrected command back to the shell history.
	AlterHistory bool `json:"alter_history"`
	// WaitSlowCommand is the timeout for commands listed in SlowCommands.
	WaitSlowCommand int `json:"wait_slow_command"`
	// SlowCommands lists commands that are expected to be slow to re-run.
	SlowCommands []string `json:"slow_commands"`
	// Repeat re-runs go-fuck if the corrected command fails again.
	Repeat bool `json:"repeat"`
	// InstantMode enables the experimental instant mode.
	InstantMode bool `json:"instant_mode"`
	// NumCloseMatches is the maximum number of suggestions for misspelled
	// commands.
	NumCloseMatches int `json:"num_close_matches"`
	// Env is merged into the environment of re-run commands.
	Env map[string]string `json:"env"`
	// ExcludedSearchPathPrefixes lists PATH prefixes that are not searched for
	// executables.
	ExcludedSearchPathPrefixes []string `json:"excluded_search_path_prefixes"`

	// UserDir is the directory that stores the user settings file.
	UserDir string `json:"-"`

	// extra holds rule specific settings that are not part of the core
	// configuration, for example "fixlinecmd".
	extra map[string]string
}

// Default returns settings populated with the built-in defaults.
func Default() *Settings {
	return &Settings{
		Rules:                      []string{consts.AllEnabled},
		ExcludeRules:               []string{},
		WaitCommand:                3,
		RequireConfirmation:        true,
		NoColors:                   false,
		Debug:                      false,
		Priority:                   map[string]int{},
		HistoryLimit:               0,
		AlterHistory:               true,
		WaitSlowCommand:            15,
		SlowCommands:               []string{"lein", "react-native", "gradle", "./gradlew", "vagrant"},
		Repeat:                     false,
		InstantMode:                false,
		NumCloseMatches:            3,
		Env:                        map[string]string{"LC_ALL": "C", "LANG": "C", "GIT_TRACE": "1"},
		ExcludedSearchPathPrefixes: []string{},
		extra:                      map[string]string{},
	}
}

var (
	currentMu sync.RWMutex
	current   = Default()
)

// Current returns the process wide settings.
func Current() *Settings {
	currentMu.RLock()
	defer currentMu.RUnlock()
	return current
}

// Reset restores the default settings. It is intended for tests.
func Reset() {
	currentMu.Lock()
	defer currentMu.Unlock()
	current = Default()
}

// Set replaces the process wide settings. It is intended for tests.
func Set(s *Settings) {
	currentMu.Lock()
	defer currentMu.Unlock()
	current = s
}

// Args lists the command line flags that influence settings.
type Args struct {
	Yes    bool
	Debug  bool
	Repeat bool
}

// String returns the value of a rule specific string setting, or def when the
// setting is unset.
func (s *Settings) String(key, def string) string {
	if s.extra == nil {
		return def
	}
	if v, ok := s.extra[key]; ok {
		return v
	}
	return def
}

// SetExtra sets a rule specific string setting. It is intended for tests and
// for the settings file loader.
func (s *Settings) SetExtra(key, value string) {
	if s.extra == nil {
		s.extra = map[string]string{}
	}
	s.extra[key] = value
}

// Alias returns the name of the alias the user is calling (normally "fuck").
func Alias() string {
	if alias := os.Getenv("TF_ALIAS"); alias != "" {
		return alias
	}
	return "fuck"
}

// Load applies the settings file, environment variables and command line
// arguments on top of the built-in defaults. It returns the populated settings
// and installs them as the process wide settings.
func Load(args *Args) (*Settings, error) {
	s := Default()
	s.UserDir = userDir()

	fileErr := s.loadFile()
	s.loadEnv()
	s.loadArgs(args)

	currentMu.Lock()
	current = s
	currentMu.Unlock()

	return s, fileErr
}

// userDir returns the go-fuck configuration directory, creating it when
// necessary.
func userDir() string {
	base, err := os.UserConfigDir()
	if err != nil || base == "" {
		home, homeErr := os.UserHomeDir()
		if homeErr != nil {
			return ".go-fuck"
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "go-fuck")
}

// fileSchema mirrors the settings file layout. Both the flattened form and the
// "settings" map for rule specific values are supported.
type fileSchema struct {
	Rules                      *[]string          `json:"rules"`
	ExcludeRules               *[]string          `json:"exclude_rules"`
	WaitCommand                *int               `json:"wait_command"`
	RequireConfirmation        *bool              `json:"require_confirmation"`
	NoColors                   *bool              `json:"no_colors"`
	Debug                      *bool              `json:"debug"`
	Priority                   *map[string]int    `json:"priority"`
	HistoryLimit               *int               `json:"history_limit"`
	AlterHistory               *bool              `json:"alter_history"`
	WaitSlowCommand            *int               `json:"wait_slow_command"`
	SlowCommands               *[]string          `json:"slow_commands"`
	Repeat                     *bool              `json:"repeat"`
	InstantMode                *bool              `json:"instant_mode"`
	NumCloseMatches            *int               `json:"num_close_matches"`
	Env                        *map[string]string `json:"env"`
	ExcludedSearchPathPrefixes *[]string          `json:"excluded_search_path_prefixes"`
	Extra                      map[string]any     `json:"settings"`
	Comment                    string             `json:"_comment"`
}

// loadFile reads the user settings file. A missing file is created with the
// default settings and is not considered an error.
func (s *Settings) loadFile() error {
	if s.UserDir == "" {
		return nil
	}
	if err := os.MkdirAll(s.UserDir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	path := filepath.Join(s.UserDir, "settings.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s.writeDefaultFile(path)
		}
		return fmt.Errorf("read settings file: %w", err)
	}

	var schema fileSchema
	if err := json.Unmarshal(data, &schema); err != nil {
		return fmt.Errorf("parse settings file %s: %w", path, err)
	}
	s.applyFile(&schema)
	return nil
}

// writeDefaultFile creates settings.json with the default configuration.
func (s *Settings) writeDefaultFile(path string) error {
	schema := map[string]any{
		"_comment":                      consts.SettingsHeader,
		"rules":                         s.Rules,
		"exclude_rules":                 s.ExcludeRules,
		"wait_command":                  s.WaitCommand,
		"require_confirmation":          s.RequireConfirmation,
		"no_colors":                     s.NoColors,
		"debug":                         s.Debug,
		"priority":                      s.Priority,
		"history_limit":                 s.HistoryLimit,
		"alter_history":                 s.AlterHistory,
		"wait_slow_command":             s.WaitSlowCommand,
		"slow_commands":                 s.SlowCommands,
		"repeat":                        s.Repeat,
		"instant_mode":                  s.InstantMode,
		"num_close_matches":             s.NumCloseMatches,
		"env":                           s.Env,
		"excluded_search_path_prefixes": s.ExcludedSearchPathPrefixes,
		"settings":                      map[string]any{},
	}
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("encode default settings: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write default settings file: %w", err)
	}
	return nil
}

// applyFile copies fields that are present in the settings file.
func (s *Settings) applyFile(f *fileSchema) {
	if f.Rules != nil {
		s.Rules = *f.Rules
	}
	if f.ExcludeRules != nil {
		s.ExcludeRules = *f.ExcludeRules
	}
	if f.WaitCommand != nil {
		s.WaitCommand = *f.WaitCommand
	}
	if f.RequireConfirmation != nil {
		s.RequireConfirmation = *f.RequireConfirmation
	}
	if f.NoColors != nil {
		s.NoColors = *f.NoColors
	}
	if f.Debug != nil {
		s.Debug = *f.Debug
	}
	if f.Priority != nil {
		s.Priority = *f.Priority
	}
	if f.HistoryLimit != nil {
		s.HistoryLimit = *f.HistoryLimit
	}
	if f.AlterHistory != nil {
		s.AlterHistory = *f.AlterHistory
	}
	if f.WaitSlowCommand != nil {
		s.WaitSlowCommand = *f.WaitSlowCommand
	}
	if f.SlowCommands != nil {
		s.SlowCommands = *f.SlowCommands
	}
	if f.Repeat != nil {
		s.Repeat = *f.Repeat
	}
	if f.InstantMode != nil {
		s.InstantMode = *f.InstantMode
	}
	if f.NumCloseMatches != nil {
		s.NumCloseMatches = *f.NumCloseMatches
	}
	if f.Env != nil {
		s.Env = *f.Env
	}
	if f.ExcludedSearchPathPrefixes != nil {
		s.ExcludedSearchPathPrefixes = *f.ExcludedSearchPathPrefixes
	}
	for key, value := range f.Extra {
		if str, ok := value.(string); ok {
			s.SetExtra(key, str)
		}
	}
}

// envPrefixes lists the accepted environment variable prefixes, in order of
// precedence.
var envPrefixes = []string{"GOFUCK_", "THEFUCK_"}

// loadEnv applies environment variables to the settings.
func (s *Settings) loadEnv() {
	for _, prefix := range envPrefixes {
		s.rulesFromEnv(prefix)
		s.excludeRulesFromEnv(prefix)
		s.intFromEnv(prefix, "WAIT_COMMAND", &s.WaitCommand)
		s.boolFromEnv(prefix, "REQUIRE_CONFIRMATION", &s.RequireConfirmation)
		s.boolFromEnv(prefix, "NO_COLORS", &s.NoColors)
		s.boolFromEnv(prefix, "DEBUG", &s.Debug)
		s.priorityFromEnv(prefix)
		s.intFromEnv(prefix, "HISTORY_LIMIT", &s.HistoryLimit)
		s.boolFromEnv(prefix, "ALTER_HISTORY", &s.AlterHistory)
		s.intFromEnv(prefix, "WAIT_SLOW_COMMAND", &s.WaitSlowCommand)
		s.sliceFromEnv(prefix, "SLOW_COMMANDS", &s.SlowCommands)
		s.boolFromEnv(prefix, "REPEAT", &s.Repeat)
		s.boolFromEnv(prefix, "INSTANT_MODE", &s.InstantMode)
		s.intFromEnv(prefix, "NUM_CLOSE_MATCHES", &s.NumCloseMatches)
		s.sliceFromEnv(prefix, "EXCLUDED_SEARCH_PATH_PREFIXES", &s.ExcludedSearchPathPrefixes)
	}
}

// envValue returns the value of the first set variable for the given suffix.
func envValue(prefix, suffix string) (string, bool) {
	value, ok := os.LookupEnv(prefix + suffix)
	return value, ok
}

// boolFromEnv applies a boolean environment variable.
func (s *Settings) boolFromEnv(prefix, suffix string, dst *bool) {
	if value, ok := envValue(prefix, suffix); ok {
		*dst = strings.EqualFold(value, "true")
	}
}

// intFromEnv applies an integer environment variable.
func (s *Settings) intFromEnv(prefix, suffix string, dst *int) {
	if value, ok := envValue(prefix, suffix); ok {
		if parsed, err := strconv.Atoi(value); err == nil {
			*dst = parsed
		}
	}
}

// sliceFromEnv applies a colon separated list environment variable.
func (s *Settings) sliceFromEnv(prefix, suffix string, dst *[]string) {
	if value, ok := envValue(prefix, suffix); ok {
		*dst = strings.Split(value, ":")
	}
}

// rulesFromEnv applies the enabled rules list.
func (s *Settings) rulesFromEnv(prefix string) {
	value, ok := envValue(prefix, "RULES")
	if !ok {
		return
	}
	rules := []string{}
	hasDefault := false
	for _, rule := range strings.Split(value, ":") {
		if rule == "DEFAULT_RULES" {
			hasDefault = true
			continue
		}
		rules = append(rules, rule)
	}
	if hasDefault {
		s.Rules = append([]string{consts.AllEnabled}, rules...)
	} else {
		s.Rules = rules
	}
}

// excludeRulesFromEnv applies the excluded rules list.
func (s *Settings) excludeRulesFromEnv(prefix string) {
	if value, ok := envValue(prefix, "EXCLUDE_RULES"); ok {
		s.ExcludeRules = strings.Split(value, ":")
	}
}

// priorityFromEnv applies rule priority overrides.
func (s *Settings) priorityFromEnv(prefix string) {
	value, ok := envValue(prefix, "PRIORITY")
	if !ok {
		return
	}
	for _, part := range strings.Split(value, ":") {
		rule, raw, found := strings.Cut(part, "=")
		if !found {
			continue
		}
		priority, err := strconv.Atoi(raw)
		if err != nil {
			continue
		}
		s.Priority[rule] = priority
	}
}

// loadArgs applies command line arguments.
func (s *Settings) loadArgs(args *Args) {
	if args == nil {
		return
	}
	if args.Yes {
		s.RequireConfirmation = false
	}
	if args.Debug {
		s.Debug = true
	}
	if args.Repeat {
		s.Repeat = true
	}
}

// ShellFromEnv returns the declared shell from TF_SHELL, if any.
func ShellFromEnv() string {
	return os.Getenv("TF_SHELL")
}
