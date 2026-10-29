// Package customrules loads user defined correction rules from JSON files.
//
// A rule file lives in the "rules" directory of the go-fuck configuration
// directory (see conf.UserDir) and has the following shape:
//
//	{
//	  "name": "git_remote_typo",
//	  "enabled": true,
//	  "priority": 900,
//	  "match": {
//	    "script": "^git remote add (\\w+)$",
//	    "output": "No such remote '([^']+)'"
//	  },
//	  "commands": ["git remote remove {1}"]
//	}
//
// Both match expressions are optional, but at least one is required. Captures
// of the output expression can be used in command templates as {0}, {1}, ...
// and the original command line as {script}. Rules are declarative: they can
// produce corrected commands but cannot run code or side effects.
package customrules

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/leafevans/go-fuck/internal/consts"
	"github.com/leafevans/go-fuck/internal/types"
)

// ruleNamePattern matches the names accepted for custom rules. It uses the
// same alphabet as the names of the compiled in rules.
var ruleNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_.]*$`)

// placeholderPattern matches the placeholders supported in command templates.
var placeholderPattern = regexp.MustCompile(`\{(script|[0-9]+)\}`)

// ruleFile mirrors the JSON layout of a custom rule.
type ruleFile struct {
	Name     string    `json:"name"`
	Enabled  *bool     `json:"enabled"`
	Priority int       `json:"priority"`
	Match    matchFile `json:"match"`
	Commands []string  `json:"commands"`
}

// matchFile mirrors the "match" object of a custom rule.
type matchFile struct {
	Script string `json:"script"`
	Output string `json:"output"`
}

// definition is a validated custom rule.
type definition struct {
	name     string
	enabled  bool
	priority int
	script   *regexp.Regexp
	output   *regexp.Regexp
	commands []string
}

// Load reads the JSON rule files from dir. It returns the valid rules and one
// error per skipped file. A missing directory is not an error.
func Load(dir string) ([]*types.Rule, []error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, []error{fmt.Errorf("read custom rules: %w", err)}
	}

	rules := []*types.Rule{}
	problems := []error{}
	seen := map[string]string{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		rule, problem := loadRule(path)
		if problem != nil {
			problems = append(problems, problem)
			continue
		}
		if previous, exists := seen[rule.Name]; exists {
			problems = append(problems, fmt.Errorf("%s: rule name %q is already used by %s", path, rule.Name, previous))
			continue
		}
		seen[rule.Name] = path
		rules = append(rules, rule)
	}
	return rules, problems
}

// loadRule reads and validates a single rule file.
func loadRule(path string) (*types.Rule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	// Windows editors often save UTF-8 with a byte order mark.
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})

	var file ruleFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	definition, err := newDefinition(&file)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return definition.rule(), nil
}

// newDefinition validates a rule file and returns its runtime representation.
func newDefinition(file *ruleFile) (*definition, error) {
	if !ruleNamePattern.MatchString(file.Name) {
		return nil, fmt.Errorf("invalid rule name %q", file.Name)
	}
	if file.Match.Script == "" && file.Match.Output == "" {
		return nil, errors.New(`at least one of "match.script" and "match.output" is required`)
	}
	if len(file.Commands) == 0 {
		return nil, errors.New(`"commands" must contain at least one command`)
	}

	priority := file.Priority
	if priority == 0 {
		priority = consts.DefaultPriority
	}
	if priority < 0 {
		return nil, fmt.Errorf("invalid priority %d", file.Priority)
	}

	enabled := true
	if file.Enabled != nil {
		enabled = *file.Enabled
	}

	script, err := compileExpression(file.Match.Script)
	if err != nil {
		return nil, fmt.Errorf(`"match.script": %w`, err)
	}
	output, err := compileExpression(file.Match.Output)
	if err != nil {
		return nil, fmt.Errorf(`"match.output": %w`, err)
	}

	groups := 0
	if output != nil {
		groups = output.NumSubexp()
	}
	commands := make([]string, 0, len(file.Commands))
	for index, command := range file.Commands {
		if strings.TrimSpace(command) == "" {
			return nil, fmt.Errorf("command %d is empty", index+1)
		}
		if err := validatePlaceholders(command, groups, output != nil); err != nil {
			return nil, fmt.Errorf("command %d: %w", index+1, err)
		}
		commands = append(commands, command)
	}

	return &definition{
		name:     file.Name,
		enabled:  enabled,
		priority: priority,
		script:   script,
		output:   output,
		commands: commands,
	}, nil
}

// compileExpression compiles a match expression, or returns nil when the
// expression is empty.
func compileExpression(expression string) (*regexp.Regexp, error) {
	if expression == "" {
		return nil, nil
	}
	return regexp.Compile(expression)
}

// validatePlaceholders reports placeholders that cannot be expanded.
func validatePlaceholders(command string, groups int, hasOutput bool) error {
	for _, match := range placeholderPattern.FindAllStringSubmatch(command, -1) {
		placeholder := match[1]
		if placeholder == "script" {
			continue
		}
		index, err := strconv.Atoi(placeholder)
		if err != nil {
			return fmt.Errorf("invalid placeholder {%s}", placeholder)
		}
		if !hasOutput {
			return fmt.Errorf("placeholder {%d} requires \"match.output\"", index)
		}
		if index > groups {
			return fmt.Errorf("placeholder {%d} exceeds the %d capture group(s) of \"match.output\"", index, groups)
		}
	}
	return nil
}

// rule builds the rule that is registered with the rules package.
func (d *definition) rule() *types.Rule {
	rule := types.NewRule(d.name, d.matches, d.newCommands)
	rule.EnabledByDefault = d.enabled
	rule.Priority = d.priority
	rule.RequiresOutput = d.output != nil
	return rule
}

// matches reports whether the rule applies to command.
func (d *definition) matches(command *types.Command) bool {
	if d.script != nil && !d.script.MatchString(command.Script) {
		return false
	}
	if d.output != nil {
		if !command.HasOutput() || !d.output.MatchString(command.Output) {
			return false
		}
	}
	return true
}

// newCommands expands the command templates for command.
func (d *definition) newCommands(command *types.Command) ([]string, error) {
	var captures []string
	if d.output != nil {
		captures = d.output.FindStringSubmatch(command.Output)
		if captures == nil {
			return nil, nil
		}
	}

	commands := make([]string, 0, len(d.commands))
	for _, template := range d.commands {
		commands = append(commands, expandPlaceholders(template, captures, command.Script))
	}
	return commands, nil
}

// expandPlaceholders replaces the supported placeholders in template.
func expandPlaceholders(template string, captures []string, script string) string {
	return placeholderPattern.ReplaceAllStringFunc(template, func(placeholder string) string {
		name := placeholder[1 : len(placeholder)-1]
		if name == "script" {
			return script
		}
		index, err := strconv.Atoi(name)
		if err != nil || index >= len(captures) {
			return placeholder
		}
		return captures[index]
	})
}
