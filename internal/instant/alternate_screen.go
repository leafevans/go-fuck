package instant

import "strings"

// alternateScreenEnter and alternateScreenLeave are the sequences that switch
// between the main and the alternate screen. Full screen programs such as vim
// or tmux use them.
var alternateScreenEnter = []string{"\x1b[?1049h", "\x1b[?1047h", "\x1b[?47h"}
var alternateScreenLeave = []string{"\x1b[?1049l", "\x1b[?1047l", "\x1b[?47l"}

// stripAlternateScreen drops the text that a full screen program wrote while
// it was active on the alternate screen. The terminal restores the main
// screen when the program exits, so that text is not part of the usable log.
// Nested alternate screens are tracked with a depth counter.
func stripAlternateScreen(data string) string {
	var builder strings.Builder
	builder.Grow(len(data))
	depth := 0
	for index := 0; index < len(data); {
		if length, matched := matchSequence(data[index:], alternateScreenEnter); matched {
			depth++
			index += length
			continue
		}
		if length, matched := matchSequence(data[index:], alternateScreenLeave); matched {
			if depth > 0 {
				depth--
			}
			index += length
			continue
		}
		if depth == 0 {
			builder.WriteByte(data[index])
		}
		index++
	}
	return builder.String()
}

// matchSequence reports whether data starts with one of sequences and returns
// the sequence length.
func matchSequence(data string, sequences []string) (int, bool) {
	for _, sequence := range sequences {
		if strings.HasPrefix(data, sequence) {
			return len(sequence), true
		}
	}
	return 0, false
}
