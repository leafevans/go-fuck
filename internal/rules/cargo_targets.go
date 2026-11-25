package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v4/internal/utils"
)

// cargoAvailableTargets returns the target names that cargo lists after the
// given section marker.
func cargoAvailableTargets(output, marker string) []string {
	_, after, ok := strings.Cut(output, marker)
	if !ok {
		return nil
	}
	var targets []string
	for line := range strings.SplitSeq(after, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(line, "    ") {
			break
		}
		targets = append(targets, trimmed)
	}
	return targets
}

// cargoClosestTarget returns the closest target to missing, or the only
// target when there is exactly one.
func cargoClosestTarget(missing string, targets []string) string {
	target := utils.GetClosest(missing, targets, 0.6, false)
	if target == "" && len(targets) == 1 {
		target = targets[0]
	}
	return target
}

func cargoGlobalOption(name string) commandOption {
	switch name {
	case "--config", "--color", "--directory", "-C", "-Z":
		return valueOption
	case "-v", "-vv", "--verbose", "-q", "--quiet", "--frozen", "--locked", "--offline":
		return flagOption
	default:
		return unknownOption
	}
}

func cargoSubcommandIndex(words []commandWord) int {
	index := 1
	if len(words) > index && strings.HasPrefix(words[index].text, "+") {
		index++
	}
	return subcommandIndex(words, index, cargoGlobalOption)
}

func cargoTargetOption(name string) commandOption {
	switch name {
	case "--bin", "--bench", "--example", "--test", "--features", "-F", "--package", "-p", "--exclude",
		"--manifest-path", "--target", "--target-dir", "--profile", "--jobs", "-j", "--message-format":
		return valueOption
	case "--release", "-r", "--workspace", "--all", "--all-features", "--no-default-features", "--no-run",
		"--all-targets", "--lib", "--bins", "--benches", "--examples", "--tests", "--doc", "--no-fail-fast":
		return flagOption
	default:
		return cargoGlobalOption(name)
	}
}

// cargoReplaceTarget edits a unique value of the specified Cargo selector,
// ignoring other option values and arguments passed to the test or binary.
func cargoReplaceTarget(script, option, missing, replacement string) []string {
	words := simpleCommandWords(script)
	index := cargoSubcommandIndex(words)
	if index < 0 {
		return nil
	}
	var target commandWord
	var prefix string
	for index++; index < len(words); {
		word := words[index]
		if word.text == "--" {
			break
		}
		if !strings.HasPrefix(word.text, "-") {
			index++
			continue
		}
		end, ok := optionEnd(words, index, cargoTargetOption)
		if !ok {
			return nil
		}
		name, value, inline := strings.Cut(word.text, "=")
		if name == option {
			candidate, candidatePrefix := word, name+"="
			if !inline {
				candidate, candidatePrefix = words[index+1], ""
				value = candidate.text
			}
			if value == missing {
				if target.end != 0 {
					return nil
				}
				target, prefix = candidate, candidatePrefix
			}
		}
		index = end
	}
	if target.end == 0 {
		return nil
	}
	return replaceCommandWord(script, target, prefix, replacement)
}
