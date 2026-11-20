// Package instant implements the instant mode.
//
// In instant mode the shell alias starts a logged shell: the shell logger
// copies the terminal output of that shell into a ring buffer, and go-fuck
// reads the output of the last command from the log instead of re-running it.
//
// The log reader and the ring buffer writer are platform independent. Spawning
// the logged shell needs a pseudo terminal and is implemented per platform
// (POSIX pseudo terminals on Linux and macOS, ConPTY on Windows). The ConPTY
// stream encodes line breaks and redraws with cursor positions, so it is
// rendered with a small screen model before the log is parsed.
package instant

import (
	"errors"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/leafevans/go-fuck/v2/internal/consts"
)

// ErrUnsupported reports that the shell logger is not available on the
// current platform.
var ErrUnsupported = errors.New("shell logger is not supported on this platform")

// LogPath returns the terminal log path of the running instant mode session.
//
// It returns an empty string when instant mode is not active.
func LogPath() string {
	if path := os.Getenv(consts.OutputLogEnv); path != "" {
		return path
	}
	return os.Getenv(consts.LegacyOutputLogEnv)
}

// LastOutput returns the output of script from the instant mode log.
//
// The second result reports whether the command was found in the log. When it
// is false the caller should fall back to re-running the command.
func LastOutput(logPath, script string) (string, bool) {
	data, err := readLog(logPath)
	if err != nil || len(data) == 0 {
		return "", false
	}

	parts := strings.Fields(script)
	calls := groupCalls(logLines(data))
	for _, call := range slices.Backward(calls) {
		echo := call.echoIndex(script, parts)
		if echo < 0 {
			continue
		}
		return call.output(echo), true
	}
	return "", false
}

// logCall is one prompt line of the terminal log together with the lines that
// followed it (the command output).
type logCall struct {
	prompt string
	lines  []string
}

// echoIndex returns the index of the prompt block line that contains the
// echoed invocation of script, or -1 when this call is a different command.
//
// Multi line prompts (for example the oh-my-zsh bira theme) put the marker on
// the first prompt line and the typed command on a later one, so the whole
// prompt block has to be searched instead of only the marker line.
func (c logCall) echoIndex(script string, parts []string) int {
	trimmed := strings.TrimSpace(script)
	if trimmed == "" {
		return -1
	}
	const maxPromptLines = 16
	lines := append([]string{c.prompt}, c.lines[:min(len(c.lines), maxPromptLines-1)]...)
	for index, line := range lines {
		cleaned := strings.TrimSpace(sanitizeEcho(line))
		if hasCommandSuffix(cleaned, trimmed) {
			return index
		}
		// Shell redraws may change argument spacing, but argument order and
		// complete words must still match the end of the echoed invocation.
		fields := strings.Fields(cleaned)
		if len(parts) > 0 && len(fields) >= len(parts) && slices.Equal(fields[len(fields)-len(parts):], parts) {
			return index
		}
	}
	if index := multilineEchoIndex(lines, trimmed); index >= 0 {
		return index
	}
	return wrappedEchoIndex(lines, trimmed)
}

func hasCommandSuffix(line, command string) bool {
	if !strings.HasSuffix(line, command) {
		return false
	}
	prefix := line[:len(line)-len(command)]
	if prefix == "" {
		return true
	}
	last, _ := utf8.DecodeLastRuneInString(prefix)
	return !unicode.IsLetter(last) && !unicode.IsDigit(last) && !strings.ContainsRune("_./\\-", last)
}

// sanitizeEcho applies backspace edits before stripping controls. Removing a
// backspace alone would join a stale typed character to the echoed command.
func sanitizeEcho(line string) string {
	if strings.ContainsRune(line, '\b') {
		runes := make([]rune, 0, len(line))
		for _, character := range line {
			if character == '\b' {
				if len(runes) > 0 {
					runes = runes[:len(runes)-1]
				}
			} else {
				runes = append(runes, character)
			}
		}
		line = string(runes)
	}
	return sanitize(line)
}

// multilineEchoIndex returns the index of the last prompt block line that
// echoes a command typed over several lines (for example PSReadLine with
// Shift+Enter, or a pasted multi line command).
//
// The search is bounded to the beginning of the prompt block, and every non
// empty command line has to appear in order, which keeps output lines from
// being mistaken for the echo.
func multilineEchoIndex(lines []string, command string) int {
	commandLines := strings.Split(command, "\n")
	if len(commandLines) < 2 {
		return -1
	}
	parts := make([]string, 0, len(commandLines))
	for _, line := range commandLines {
		line = strings.TrimSpace(line)
		if line != "" {
			parts = append(parts, line)
		}
	}
	if len(parts) < 2 {
		return -1
	}
	const maxPromptLines = 16
	last := min(maxPromptLines, len(lines))
	for start := range last {
		index := start
		matched := 0
		lastMatch := -1
		for _, part := range parts {
			found := -1
			for ; index < last; index++ {
				if strings.Contains(sanitizeEcho(lines[index]), part) {
					found = index
					index++
					break
				}
			}
			if found < 0 {
				break
			}
			matched++
			lastMatch = found
		}
		if matched == len(parts) {
			return lastMatch
		}
	}
	return -1
}

// instantMaxEchoLines bounds the number of prompt block lines that are joined
// when a long command is wrapped at the terminal width.
const instantMaxEchoLines = 8

// wrappedEchoIndex finds a command that the terminal wrapped over several
// lines. The joined text is only trusted when the command ends exactly at the
// end of its line, so that an output line which merely contains the command is
// not mistaken for the echoed command.
func wrappedEchoIndex(lines []string, command string) int {
	if command == "" || len(lines) < 2 {
		return -1
	}
	var joined strings.Builder
	ends := make([]int, 0, instantMaxEchoLines)
	for index, line := range lines {
		if index >= instantMaxEchoLines {
			break
		}
		joined.WriteString(sanitizeEcho(line))
		ends = append(ends, joined.Len())
	}
	text := joined.String()
	offset := strings.Index(text, command)
	if offset < 0 || !hasCommandSuffix(text[:offset+len(command)], command) {
		return -1
	}
	end := offset + len(command)
	for index, lineEnd := range ends {
		if lineEnd < end {
			continue
		}
		if strings.TrimSpace(text[end:lineEnd]) != "" {
			return -1
		}
		return index
	}
	return -1
}

// output renders the terminal screen of the command output.
//
// The reference implementation emulates a terminal with pyte; this simplified
// renderer strips escape sequences and applies carriage return overwrites. It
// is a documented deviation in README.md.
//
// from is the index of the echoed command line inside the prompt block; the
// output starts after it.
func (c logCall) output(from int) string {
	lines := make([]string, 0, len(c.lines))
	for _, line := range c.lines[from:] {
		line = strings.TrimRight(line, "\r")
		if position := strings.LastIndexByte(line, '\r'); position >= 0 {
			line = line[position+1:]
		}
		line = sanitize(line)
		lines = append(lines, strings.TrimRight(line, " \t"))
	}
	for len(lines) > 0 && lines[0] == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}

// groupCalls splits the rendered log lines into command calls at the prompt
// marker lines.
func groupCalls(lines []string) []logCall {
	calls := []logCall{}
	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		if strings.Contains(line, consts.UserCommandMark) {
			calls = append(calls, logCall{prompt: line})
			continue
		}
		if len(calls) > 0 {
			calls[len(calls)-1].lines = append(calls[len(calls)-1].lines, line)
		}
	}
	return calls
}

// readLog reads the newest part of the instant mode log.
func readLog(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if info, err := file.Stat(); err == nil && info.Size() > consts.LogSizeInBytes {
		if _, err := file.Seek(info.Size()-consts.LogSizeInBytes, io.SeekStart); err != nil {
			return "", err
		}
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

var (
	// ansiCSI matches control sequence introducer sequences such as colors.
	ansiCSI = regexp.MustCompile(`\x1b[[][0-9;?]*[ -/]*[@-~]`)
	// ansiControl matches the remaining zero width control characters. Tabs
	// are kept because rules may look at indentation.
	ansiControl = regexp.MustCompile(`[\x00-\x08\x0b-\x1f\x7f]`)
)

// sanitize removes the prompt marker, escape sequences and control characters
// from a terminal log line.
func sanitize(line string) string {
	line = strings.ReplaceAll(line, consts.UserCommandMark, "")
	line = stripOSC(line)
	line = ansiCSI.ReplaceAllString(line, "")
	line = ansiControl.ReplaceAllString(line, "")
	// Removing an escape or control sequence can join the fragments of a
	// prompt marker that the sequence split (for example, one marker rune,
	// a DEL byte and nine marker runes). Remove the marker once more after the
	// stripping pass so the rendered output never leaks it.
	return strings.ReplaceAll(line, consts.UserCommandMark, "")
}

// stripOSC removes operating system command sequences (window titles and
// similar), which are terminated by BEL or ST (ESC followed by a backslash).
func stripOSC(line string) string {
	var b strings.Builder
	inOSC := false
	for index := 0; index < len(line); index++ {
		character := line[index]
		if inOSC {
			if character == 7 {
				inOSC = false
			} else if character == 27 && index+1 < len(line) && line[index+1] == '\\' {
				inOSC = false
				index++
			}
			continue
		}
		if character == 27 && index+1 < len(line) && line[index+1] == ']' {
			inOSC = true
			index++
			continue
		}
		b.WriteByte(character)
	}
	return b.String()
}

// RingWriter writes the terminal log while keeping it below the configured
// size. The oldest bytes are dropped when the log becomes too large, matching
// the ring buffer of the reference implementation.
type RingWriter struct {
	file *os.File
}

// NewRingWriter returns a writer that stores at most consts.LogSizeInBytes
// bytes in file.
func NewRingWriter(file *os.File) *RingWriter {
	return &RingWriter{file: file}
}

// Write appends p to the log, dropping the oldest bytes when needed.
func (w *RingWriter) Write(p []byte) (int, error) {
	if w.file == nil || len(p) == 0 {
		return len(p), nil
	}
	// A single write may be larger than the entire ring. Consume all input
	// while persisting only its newest bytes, as with repeated small writes.
	dropped := max(0, len(p)-consts.LogSizeInBytes)
	p = p[dropped:]
	info, err := w.file.Stat()
	if err != nil {
		return 0, err
	}
	if info.Size()+int64(len(p)) > consts.LogSizeInBytes {
		keep := int64(consts.LogSizeInBytes - consts.LogSizeToClean)
		if keep+int64(len(p)) > consts.LogSizeInBytes {
			keep = int64(consts.LogSizeInBytes) - int64(len(p))
		}
		if err := trimFile(w.file, keep); err != nil {
			return 0, err
		}
	}
	n, err := w.file.Write(p)
	if err != nil {
		return n, err
	}
	return dropped + n, nil
}

// trimFile keeps the last keep bytes of file.
func trimFile(file *os.File, keep int64) error {
	var tail []byte
	if keep > 0 {
		if _, err := file.Seek(-keep, io.SeekEnd); err != nil {
			return err
		}
		data, err := io.ReadAll(file)
		if err != nil {
			return err
		}
		tail = data
	}
	if err := file.Truncate(0); err != nil {
		return err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if len(tail) == 0 {
		return nil
	}
	_, err := file.Write(tail)
	return err
}
