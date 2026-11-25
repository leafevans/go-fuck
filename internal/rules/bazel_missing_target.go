package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v4/internal/types"
)

// Bazel used to put a quoted suggestion before "defined by". Modern Bazel
// appends a bare name or a list after the BUILD file path.
var (
	bazelMissingTargetError   = newRegexp("^ERROR: (?:Skipping '([^']+)': )?no such target '([^']+)': target '([^']+)' not declared in package '([^']*)' (.+)$")
	bazelMissingTargetOldHint = newRegexp("^\\(did you mean '([^']+)'\\?\\) defined by .+$")
	bazelMissingTargetHint    = newRegexp("^defined by .+? \\(did you mean ([^()]+)\\?\\)\\.?$")
	bazelMissingTargetName    = newRegexp("^[A-Za-z0-9_.+-]+(?:/[A-Za-z0-9_.+-]+)*$")
	// Recognize only simple shell words, retaining their byte offsets. This
	// deliberately excludes compound commands, escapes and shell expansions.
	bazelMissingTargetWord = newRegexp("(?:[A-Za-z0-9_@/:.,+=-]+|'[^'\\r\\n]*'|\"[^\"$\\x60\\\\\\r\\n]*\")+")
)

// bazelMissingTargetParse returns the exact failed local label and Bazel's
// suggestions with the same package. Bare suggestions are target names, not
// package names or filesystem paths to resolve.
func bazelMissingTargetParse(line string) (string, []string) {
	m := bazelMissingTargetError.FindStringSubmatch(line)
	if m == nil || m[1] != "" && m[1] != m[2] || strings.Contains(m[5], " and referenced by ") {
		return "", nil
	}
	pkg, name, ok := strings.Cut(m[2], ":")
	if !ok || !strings.HasPrefix(pkg, "//") || pkg[2:] != m[4] || name != m[3] ||
		(pkg != "//" && !bazelMissingTargetSafeName(pkg[2:])) || !bazelMissingTargetSafeName(name) {
		return "", nil
	}

	var hints []string
	if old := bazelMissingTargetOldHint.FindStringSubmatch(m[5]); old != nil {
		hints = []string{old[1]}
	} else if modern := bazelMissingTargetHint.FindStringSubmatch(m[5]); modern != nil {
		// TargetSuggester joins list entries as "a, b, or c".
		first, last, list := strings.Cut(modern[1], ", or ")
		if list {
			hints = append(strings.Split(first, ", "), last)
		} else {
			hints = []string{first}
		}
	}

	seen := map[string]bool{name: true}
	var labels []string
	for _, hint := range hints {
		if hintPkg, hintName, qualified := strings.Cut(hint, ":"); qualified {
			if hintPkg != "" && hintPkg != pkg {
				continue
			}
			hint = hintName
		}
		if bazelMissingTargetSafeName(hint) && !seen[hint] {
			seen[hint] = true
			labels = append(labels, pkg+":"+hint)
		}
	}
	return m[2], labels
}

func bazelMissingTargetSafeName(name string) bool {
	if !bazelMissingTargetName.MatchString(name) || name == "all" || name == "all-targets" {
		return false
	}
	for part := range strings.SplitSeq(name, "/") {
		if part == "." || part == ".." || part == "..." {
			return false
		}
	}
	return true
}

// bazelMissingTargetPosition finds a unique, literal target operand. A label
// immediately following a flag without '=' might be the flag's separate value;
// without an option schema we leave that command alone.
func bazelMissingTargetPosition(script, label string) (int, int, bool) {
	spans := bazelMissingTargetWord.expression().FindAllStringIndex(script, -1)
	prev := 0
	for i, span := range spans {
		if (i > 0 && prev == span[0]) || strings.Trim(script[prev:span[0]], " \t") != "" {
			return 0, 0, false
		}
		prev = span[1]
	}
	if strings.Trim(script[prev:], " \t") != "" || len(spans) < 3 {
		return 0, 0, false
	}

	word := func(span []int) (string, int) {
		start, end := span[0], span[1]
		if q := script[start]; end-start >= 2 && (q == '\'' || q == '"') && script[end-1] == q {
			start++
			end--
		}
		return script[start:end], start
	}
	app, _ := word(spans[0])
	if app != "bazel" && app != "bazel.exe" {
		return 0, 0, false
	}

	inTargets, bareFlag := false, false
	position := -1
	for _, span := range spans[1:] {
		arg, start := word(span)
		if arg == "--" {
			break
		}
		if !inTargets {
			switch arg {
			case "build", "test", "run", "coverage":
				inTargets = true
			default:
				if !strings.HasPrefix(arg, "--") {
					return 0, 0, false
				}
			}
			continue
		}
		if arg == label {
			if position >= 0 || bareFlag {
				return 0, 0, false
			}
			position = start
		}
		bareFlag = strings.HasPrefix(arg, "-") && !strings.Contains(arg, "=")
	}
	return position, position + len(label), position >= 0
}

func bazelMissingTargetCommands(command *types.Command) []string {
	if !strings.Contains(command.Output, "no such target ") {
		return nil
	}
	var corrected []string
	var operand string
	seen := make(map[string]bool)
	for line := range strings.SplitSeq(command.Output, "\n") {
		label, hints := bazelMissingTargetParse(strings.TrimSpace(line))
		if len(hints) == 0 {
			continue
		}
		start, end, ok := bazelMissingTargetPosition(command.Script, label)
		if !ok {
			continue
		}
		if operand != "" && operand != label {
			return nil
		}
		operand = label
		for _, hint := range hints {
			if !seen[hint] {
				seen[hint] = true
				corrected = append(corrected, command.Script[:start]+hint+command.Script[end:])
			}
		}
	}
	return corrected
}

func bazelMissingTargetMatch(command *types.Command) bool {
	return len(bazelMissingTargetCommands(command)) != 0
}

func bazelMissingTargetNewCommand(command *types.Command) ([]string, error) {
	return bazelMissingTargetCommands(command), nil
}

func init() {
	register(types.NewRule("bazel_missing_target", bazelMissingTargetMatch, bazelMissingTargetNewCommand))
}
