package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/shells"
)

type commandWord struct {
	text       string
	start, end int
}

// simpleCommandWords rejects compound or incomplete commands: their output
// does not reliably identify which invocation failed.
func simpleCommandWords(script string) []commandWord {
	shell := shells.Current()
	var words []commandWord
	for pos := 0; pos < len(script); {
		switch script[pos] {
		case ' ', '\t':
			pos++
			continue
		case '#':
			if strings.ContainsAny(script[pos:], "\r\n") {
				return nil
			}
			return words
		}
		word, size := shells.ReadWord(script[pos:], shell)
		if size == 0 {
			return nil
		}
		words = append(words, commandWord{text: word, start: pos, end: pos + size})
		pos += size
	}
	return words
}

type commandOption uint8

const (
	unknownOption commandOption = iota
	flagOption
	valueOption
)

// optionEnd returns the index after an option and any separate value. An
// unsupported option or missing value makes the parse ambiguous.
func optionEnd(words []commandWord, index int, options func(string) commandOption) (int, bool) {
	name, _, inline := strings.Cut(words[index].text, "=")
	switch options(name) {
	case flagOption:
		return index + 1, !inline
	case valueOption:
		if inline {
			return index + 1, true
		}
		return index + 2, index+1 < len(words) && !strings.HasPrefix(words[index+1].text, "-")
	default:
		return 0, false
	}
}

func subcommandIndex(words []commandWord, index int, options func(string) commandOption) int {
	for index < len(words) {
		word := words[index].text
		switch {
		case word == "", word == "-", word == "--":
			return -1
		case !strings.HasPrefix(word, "-"):
			return index
		}
		var ok bool
		index, ok = optionEnd(words, index, options)
		if !ok {
			return -1
		}
	}
	return -1
}

func replaceCommandWord(script string, word commandWord, prefix string, replacements ...string) []string {
	shell := shells.Current()
	var results []string
	for _, replacement := range replacements {
		if replacement != "" && prefix+replacement != word.text {
			results = append(results, script[:word.start]+prefix+shell.Quote(replacement)+script[word.end:])
		}
	}
	return results
}

func replaceSubcommand(script string, words []commandWord, index int, expected string, replacements ...string) []string {
	if index < 0 || index >= len(words) || words[index].text != expected {
		return nil
	}
	return replaceCommandWord(script, words[index], "", replacements...)
}
