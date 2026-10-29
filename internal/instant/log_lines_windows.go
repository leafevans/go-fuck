//go:build windows

package instant

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// conPTYSizeRecord is the magic prefix of a terminal size record. The Windows
// shell logger writes it at the beginning of the instant mode log and after
// every resize, before the ConPTY output stream continues.
const conPTYSizeRecord = "\x00gofuck-conpty"

// defaultRenderColumns and defaultRenderRows are used when the log does not
// start with a size record, for example because the ring buffer dropped it.
const (
	defaultRenderColumns = 120
	defaultRenderRows    = 30
)

// logLines renders the ConPTY terminal stream into text lines.
//
// ConPTY does not only write printable text: new lines are often encoded as
// absolute cursor positions, PSReadLine redraws the prompt with cursor moves
// and erase sequences, and scrolling moves the viewport. Rendering the stream
// into a small screen model is therefore required to recover the text lines
// that the log reader expects.
func logLines(data string) []string {
	terminal := newTerminal(defaultRenderColumns, defaultRenderRows)
	for index := 0; index < len(data); {
		if strings.HasPrefix(data[index:], conPTYSizeRecord) {
			end := strings.IndexByte(data[index:], '\n')
			if end < 0 {
				break
			}
			fields := strings.Fields(data[index+len(conPTYSizeRecord) : index+end])
			if len(fields) == 2 {
				columns, columnErr := strconv.Atoi(fields[0])
				rows, rowErr := strconv.Atoi(fields[1])
				if columnErr == nil && rowErr == nil {
					terminal.resize(columns, rows)
				}
			}
			index += end + 1
			continue
		}

		character := data[index]
		switch {
		case character == 0x1b:
			index = terminal.escape(data, index)
		case character == '\r':
			terminal.cursorX = 0
			index++
		case character == '\n':
			// ConPTY renders new lines as CR LF, but a bare LF from the
			// stream also starts a new line; treating it like CR LF keeps
			// text written right after it at the beginning of the row.
			terminal.cursorX = 0
			terminal.lineFeed()
			index++
		case character == '\b':
			if terminal.cursorX > 0 {
				terminal.cursorX--
			}
			index++
		case character == '\t':
			terminal.tab()
			index++
		case character < 0x20 || character == 0x7f:
			index++
		default:
			character, size := utf8.DecodeRuneInString(data[index:])
			if character == utf8.RuneError && size == 1 {
				index++
				continue
			}
			terminal.put(character)
			index += size
		}
	}
	return terminal.lines()
}

// screenLine is one row of the rendered screen.
type screenLine []rune

// terminal is the minimal screen model used to render the ConPTY output.
type terminal struct {
	columns int
	screen  []screenLine
	history []string

	cursorX int
	cursorY int

	savedX int
	savedY int
}

// newTerminal returns an empty terminal with the given size.
func newTerminal(columns, rows int) *terminal {
	if columns < 1 {
		columns = defaultRenderColumns
	}
	if rows < 1 {
		rows = defaultRenderRows
	}
	return &terminal{columns: columns, screen: make([]screenLine, rows)}
}

// resize changes the size of the terminal. Rows that no longer fit are moved
// to the scrollback.
func (t *terminal) resize(columns, rows int) {
	if columns < 1 || rows < 1 {
		return
	}
	t.columns = columns
	for len(t.screen) > rows {
		t.history = append(t.history, lineText(t.screen[0]))
		t.screen = t.screen[1:]
	}
	for len(t.screen) < rows {
		t.screen = append(t.screen, nil)
	}
	if t.cursorY >= rows {
		t.cursorY = rows - 1
	}
	if t.cursorX >= columns {
		t.cursorX = columns - 1
	}
}

// put writes a printable character at the cursor and advances it.
func (t *terminal) put(character rune) {
	if t.cursorX >= t.columns {
		t.cursorX = 0
		t.lineFeed()
	}
	line := t.screen[t.cursorY]
	for len(line) <= t.cursorX {
		line = append(line, ' ')
	}
	line[t.cursorX] = character
	t.screen[t.cursorY] = line
	t.cursorX++
}

// tab advances the cursor to the next tab stop.
func (t *terminal) tab() {
	next := (t.cursorX/8 + 1) * 8
	for t.cursorX < next {
		t.put(' ')
	}
}

// lineFeed moves the cursor one line down and scrolls when needed.
func (t *terminal) lineFeed() {
	t.cursorY++
	if t.cursorY < len(t.screen) {
		return
	}
	t.scroll()
	t.cursorY = len(t.screen) - 1
}

// scroll moves the top line into the scrollback.
func (t *terminal) scroll() {
	if len(t.screen) == 0 {
		return
	}
	t.history = append(t.history, lineText(t.screen[0]))
	copy(t.screen, t.screen[1:])
	t.screen[len(t.screen)-1] = nil
}

// lines returns the scrollback followed by the visible screen.
func (t *terminal) lines() []string {
	lines := make([]string, 0, len(t.history)+len(t.screen))
	lines = append(lines, t.history...)
	for _, line := range t.screen {
		lines = append(lines, lineText(line))
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// lineText returns the text of a screen line without the trailing blanks.
func lineText(line screenLine) string {
	return strings.TrimRight(string(line), " ")
}

// escape consumes an escape sequence starting at index and returns the index
// of the next character after it.
func (t *terminal) escape(data string, index int) int {
	index++
	if index >= len(data) {
		return index
	}
	switch data[index] {
	case '[':
		index++
		start := index
		for index < len(data) && data[index] >= 0x20 && data[index] <= 0x3f {
			index++
		}
		if index >= len(data) {
			return index
		}
		t.control(data[start:index], data[index])
		return index + 1
	case ']':
		index++
		for index < len(data) {
			switch data[index] {
			case 0x07:
				return index + 1
			case 0x1b:
				if index+1 < len(data) && data[index+1] == '\\' {
					return index + 2
				}
			}
			index++
		}
		return index
	case '7':
		t.savedX, t.savedY = t.cursorX, t.cursorY
		return index + 1
	case '8':
		t.cursorX = clamp(t.savedX, 0, t.columns-1)
		t.cursorY = clamp(t.savedY, 0, len(t.screen)-1)
		return index + 1
	case 'M':
		if t.cursorY > 0 {
			t.cursorY--
			return index + 1
		}
		copy(t.screen[1:], t.screen[:len(t.screen)-1])
		t.screen[0] = nil
		return index + 1
	case 'c':
		for row := range t.screen {
			t.screen[row] = nil
		}
		t.cursorX, t.cursorY = 0, 0
		return index + 1
	default:
		return index + 1
	}
}

// control applies one control sequence.
func (t *terminal) control(raw string, final byte) {
	private := len(raw) > 0 && (raw[0] == '?' || raw[0] == '>' || raw[0] == '!' || raw[0] == '=')
	if private {
		raw = raw[1:]
	}
	numbers := parseParameters(raw)

	switch final {
	case 'A':
		t.cursorY = clamp(t.cursorY-parameter(numbers, 0, 1), 0, len(t.screen)-1)
	case 'B':
		t.cursorY = clamp(t.cursorY+parameter(numbers, 0, 1), 0, len(t.screen)-1)
	case 'C':
		t.cursorX = clamp(t.cursorX+parameter(numbers, 0, 1), 0, t.columns-1)
	case 'D':
		t.cursorX = clamp(t.cursorX-parameter(numbers, 0, 1), 0, t.columns-1)
	case 'E':
		t.cursorX, t.cursorY = 0, clamp(t.cursorY+parameter(numbers, 0, 1), 0, len(t.screen)-1)
	case 'F':
		t.cursorX, t.cursorY = 0, clamp(t.cursorY-parameter(numbers, 0, 1), 0, len(t.screen)-1)
	case 'G':
		t.cursorX = clamp(parameter(numbers, 0, 1)-1, 0, t.columns-1)
	case 'd':
		t.cursorY = clamp(parameter(numbers, 0, 1)-1, 0, len(t.screen)-1)
	case 'H', 'f':
		t.cursorY = clamp(parameter(numbers, 0, 1)-1, 0, len(t.screen)-1)
		t.cursorX = clamp(parameter(numbers, 1, 1)-1, 0, t.columns-1)
	case 'J':
		t.eraseDisplay(parameter(numbers, 0, 0))
	case 'K':
		t.eraseLine(parameter(numbers, 0, 0))
	case 'X':
		t.eraseCharacters(parameter(numbers, 0, 1))
	case 'P':
		t.deleteCharacters(parameter(numbers, 0, 1))
	case '@':
		t.insertCharacters(parameter(numbers, 0, 1))
	case 'L':
		t.insertLines(parameter(numbers, 0, 1))
	case 'M':
		t.deleteLines(parameter(numbers, 0, 1))
	case 'S':
		t.scrollUp(parameter(numbers, 0, 1))
	case 'T':
		t.scrollDown(parameter(numbers, 0, 1))
	}
}

// eraseDisplay applies an ED sequence.
func (t *terminal) eraseDisplay(mode int) {
	switch mode {
	case 0:
		t.truncateLine(t.cursorY, t.cursorX)
		for row := t.cursorY + 1; row < len(t.screen); row++ {
			t.screen[row] = nil
		}
	case 1:
		for row := 0; row < t.cursorY; row++ {
			t.screen[row] = nil
		}
		t.blankLine(t.cursorY, t.cursorX)
	case 2, 3:
		for row := range t.screen {
			t.screen[row] = nil
		}
	}
}

// eraseLine applies an EL sequence.
func (t *terminal) eraseLine(mode int) {
	switch mode {
	case 0:
		t.truncateLine(t.cursorY, t.cursorX)
	case 1:
		t.blankLine(t.cursorY, t.cursorX)
	case 2:
		t.screen[t.cursorY] = nil
	}
}

// truncateLine drops the text of row after column.
func (t *terminal) truncateLine(row, column int) {
	if row < 0 || row >= len(t.screen) {
		return
	}
	line := t.screen[row]
	if column < len(line) {
		t.screen[row] = line[:column]
	}
}

// blankLine blanks the text of row up to column.
func (t *terminal) blankLine(row, column int) {
	if row < 0 || row >= len(t.screen) {
		return
	}
	line := t.screen[row]
	for index := 0; index <= column && index < len(line); index++ {
		line[index] = ' '
	}
}

// eraseCharacters blanks count characters at the cursor.
func (t *terminal) eraseCharacters(count int) {
	if t.cursorY < 0 || t.cursorY >= len(t.screen) || count < 1 {
		return
	}
	line := t.screen[t.cursorY]
	for index := t.cursorX; index < t.cursorX+count && index < len(line); index++ {
		line[index] = ' '
	}
}

// deleteCharacters removes count characters at the cursor.
func (t *terminal) deleteCharacters(count int) {
	if t.cursorY < 0 || t.cursorY >= len(t.screen) || count < 1 {
		return
	}
	line := t.screen[t.cursorY]
	if t.cursorX >= len(line) {
		return
	}
	end := t.cursorX + count
	if end > len(line) {
		end = len(line)
	}
	t.screen[t.cursorY] = append(line[:t.cursorX], line[end:]...)
}

// insertCharacters inserts count blanks at the cursor.
func (t *terminal) insertCharacters(count int) {
	if t.cursorY < 0 || t.cursorY >= len(t.screen) || count < 1 || t.cursorX >= len(t.screen[t.cursorY]) {
		return
	}
	line := t.screen[t.cursorY]
	blanks := make(screenLine, count)
	for index := range blanks {
		blanks[index] = ' '
	}
	t.screen[t.cursorY] = append(append(line[:t.cursorX], blanks...), line[t.cursorX:]...)
}

// insertLines inserts count blank lines at the cursor.
func (t *terminal) insertLines(count int) {
	rows := len(t.screen)
	if count < 1 || t.cursorY < 0 || t.cursorY >= rows {
		return
	}
	if count > rows-t.cursorY {
		count = rows - t.cursorY
	}
	copy(t.screen[t.cursorY+count:], t.screen[t.cursorY:rows-count])
	for index := 0; index < count; index++ {
		t.screen[t.cursorY+index] = nil
	}
}

// deleteLines removes count lines at the cursor.
func (t *terminal) deleteLines(count int) {
	rows := len(t.screen)
	if count < 1 || t.cursorY < 0 || t.cursorY >= rows {
		return
	}
	if count > rows-t.cursorY {
		count = rows - t.cursorY
	}
	copy(t.screen[t.cursorY:], t.screen[t.cursorY+count:])
	for index := rows - count; index < rows; index++ {
		t.screen[index] = nil
	}
}

// scrollUp moves count lines of the viewport into the scrollback.
func (t *terminal) scrollUp(count int) {
	for index := 0; index < count; index++ {
		t.scroll()
	}
}

// scrollDown moves the viewport content down, dropping the bottom lines.
func (t *terminal) scrollDown(count int) {
	for index := 0; index < count && len(t.screen) > 0; index++ {
		copy(t.screen[1:], t.screen[:len(t.screen)-1])
		t.screen[0] = nil
	}
}

// parseParameters parses the parameters of a control sequence.
func parseParameters(raw string) []int {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ";")
	numbers := make([]int, len(parts))
	for index, part := range parts {
		if value, err := strconv.Atoi(part); err == nil {
			numbers[index] = value
		}
	}
	return numbers
}

// parameter returns the index-th parameter, or fallback when it is absent or
// zero.
func parameter(numbers []int, index, fallback int) int {
	if index >= len(numbers) || numbers[index] == 0 {
		return fallback
	}
	return numbers[index]
}

// clamp limits value to the inclusive range.
func clamp(value, minimum, maximum int) int {
	if value < minimum {
		return minimum
	}
	if maximum < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}
