//go:build windows

package instant

import (
	"strconv"
	"strings"
	"unicode"
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

	// wrapped marks the rows whose text continues on the next row because the
	// line was wider than the terminal. historyWrapped is the same flag for the
	// rows that moved into the scrollback.
	wrapped        []bool
	historyWrapped []bool

	cursorX int
	cursorY int

	savedX int
	savedY int

	// alternate keeps the main screen while a full screen program such as vim
	// or tmux is active, so that its screen contents do not corrupt the log.
	alternate *alternateScreen
}

// alternateScreen is the saved main screen of an alternate screen session.
type alternateScreen struct {
	screen            []screenLine
	wrapped           []bool
	historyLen        int
	historyWrappedLen int
	cursorX           int
	cursorY           int
}

// newTerminal returns an empty terminal with the given size.
func newTerminal(columns, rows int) *terminal {
	if columns < 1 {
		columns = defaultRenderColumns
	}
	if rows < 1 {
		rows = defaultRenderRows
	}
	return &terminal{
		columns: columns,
		screen:  make([]screenLine, rows),
		wrapped: make([]bool, rows),
	}
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
		t.historyWrapped = append(t.historyWrapped, t.wrapped[0])
		t.screen = t.screen[1:]
		t.wrapped = t.wrapped[1:]
	}
	for len(t.screen) < rows {
		t.screen = append(t.screen, nil)
		t.wrapped = append(t.wrapped, false)
	}
	if t.cursorY >= rows {
		t.cursorY = rows - 1
	}
	if t.cursorX >= columns {
		t.cursorX = columns - 1
	}
}

// continuationCell marks the second column of a wide (double width)
// character. It is never part of the rendered text.
const continuationCell rune = 0

// wideRanges lists the East Asian wide and fullwidth rune ranges that occupy
// two terminal columns.
var wideRanges = [][2]rune{
	{0x1100, 0x115F},
	{0x2329, 0x232A},
	{0x2E80, 0x303E},
	{0x3041, 0x33FF},
	{0x3400, 0x4DBF},
	{0x4E00, 0x9FFF},
	{0xA000, 0xA4CF},
	{0xA960, 0xA97F},
	{0xAC00, 0xD7A3},
	{0xF900, 0xFAFF},
	{0xFE10, 0xFE19},
	{0xFE30, 0xFE6F},
	{0xFF00, 0xFF60},
	{0xFFE0, 0xFFE6},
	{0x1B000, 0x1B2FF},
	{0x1F300, 0x1F64F},
	{0x1F680, 0x1F6FF},
	{0x1F900, 0x1F9FF},
	{0x1FA70, 0x1FAFF},
	{0x20000, 0x2FFFD},
	{0x30000, 0x3FFFD},
}

// runeWidth returns the number of terminal columns that character occupies.
func runeWidth(character rune) int {
	// The zero width space (0x200B) is part of the user command mark that the
	// log reader looks for, so it keeps its own cell instead of being dropped.
	if character < 0x20 || character == 0x200C || character == 0x200D ||
		character == 0xFEFF {
		return 0
	}
	if unicode.In(character, unicode.Mn, unicode.Me) {
		return 0
	}
	for _, wide := range wideRanges {
		if character >= wide[0] && character <= wide[1] {
			return 2
		}
	}
	return 1
}

// put writes a printable character at the cursor and advances it.
//
// Wide characters occupy two columns: the character itself and a continuation
// cell, so cursor positions after CJK or emoji text line up with what the
// terminal actually displayed.
func (t *terminal) put(character rune) {
	width := runeWidth(character)
	if width == 0 {
		return
	}
	if t.cursorX >= t.columns || width == 2 && t.cursorX+1 >= t.columns {
		t.wrapped[t.cursorY] = true
		t.cursorX = 0
		t.lineFeed()
	}
	line := ensureCells(t.screen[t.cursorY], t.cursorX+width)
	blankCell(line, t.cursorX)
	if width == 2 {
		blankCell(line, t.cursorX+1)
	}
	line[t.cursorX] = character
	if width == 2 {
		line[t.cursorX+1] = continuationCell
	}
	t.screen[t.cursorY] = line
	t.cursorX += width
}

// ensureCells grows line to at least length cells.
func ensureCells(line screenLine, length int) screenLine {
	for len(line) < length {
		line = append(line, ' ')
	}
	return line
}

// blankCell blanks one cell and repairs the other half of a wide character
// that would otherwise be orphaned.
func blankCell(line screenLine, index int) {
	if index < 0 || index >= len(line) {
		return
	}
	if line[index] == continuationCell {
		if index > 0 {
			line[index-1] = ' '
		}
		line[index] = ' '
		return
	}
	if runeWidth(line[index]) == 2 && index+1 < len(line) {
		line[index+1] = ' '
	}
	line[index] = ' '
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
	t.historyWrapped = append(t.historyWrapped, t.wrapped[0])
	copy(t.screen, t.screen[1:])
	t.screen[len(t.screen)-1] = nil
	copy(t.wrapped, t.wrapped[1:])
	t.wrapped[len(t.wrapped)-1] = false
}

// lines returns the scrollback followed by the visible screen.
//
// Rows that the screen model wrapped because they were wider than the terminal
// are joined again: the rules match the text of a command or an error message,
// not the physical rows of the terminal.
func (t *terminal) lines() []string {
	lines := make([]string, 0, len(t.history)+len(t.screen))
	wrapped := make([]bool, 0, len(t.history)+len(t.screen))
	lines = append(lines, t.history...)
	wrapped = append(wrapped, t.historyWrapped...)
	for _, line := range t.screen {
		lines = append(lines, lineText(line))
	}
	wrapped = append(wrapped, t.wrapped...)

	joined := make([]string, 0, len(lines))
	for index := 0; index < len(lines); index++ {
		text := lines[index]
		for index+1 < len(lines) && wrapsToNextRow(lines[index], wrapped[index], t.columns) {
			index++
			text += lines[index]
		}
		joined = append(joined, text)
	}
	for len(joined) > 0 && joined[len(joined)-1] == "" {
		joined = joined[:len(joined)-1]
	}
	return joined
}

// cellWidth returns the number of terminal columns that text occupies.
func cellWidth(text string) int {
	width := 0
	for _, character := range text {
		width += runeWidth(character)
	}
	return width
}

// wrapsToNextRow reports whether the rendered row continues on the next row.
//
// The screen model knows when it wrapped a row itself, but ConPTY re-emits the
// wrapped text of the child console with an explicit line break and often
// leaves the last column empty, so a row that is one column short of the
// terminal width also continues on the next row.
func wrapsToNextRow(line string, wrapped bool, columns int) bool {
	if wrapped {
		return true
	}
	return columns > 1 && cellWidth(line) >= columns-1
}

// lineText returns the text of a screen line without the continuation cells
// of wide characters and the trailing blanks.
func lineText(line screenLine) string {
	var b strings.Builder
	b.Grow(len(line))
	for _, character := range line {
		if character == continuationCell {
			continue
		}
		b.WriteRune(character)
	}
	return strings.TrimRight(b.String(), " ")
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

	if private && (final == 'h' || final == 'l') {
		t.setPrivateModes(numbers, final == 'h')
		return
	}

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

// setPrivateModes applies the private mode parameters that the screen model
// understands. Other private modes are ignored.
func (t *terminal) setPrivateModes(numbers []int, enabled bool) {
	for _, mode := range numbers {
		switch mode {
		case 47, 1047, 1049:
			if enabled {
				t.enterAlternateScreen()
			} else {
				t.leaveAlternateScreen()
			}
		}
	}
}

// enterAlternateScreen saves the main screen and starts a blank alternate
// screen. Full screen programs run there; their content is discarded when they
// exit, like the terminal restores the main screen.
func (t *terminal) enterAlternateScreen() {
	if t.alternate != nil {
		return
	}
	t.alternate = &alternateScreen{
		screen:            t.screen,
		wrapped:           t.wrapped,
		historyLen:        len(t.history),
		historyWrappedLen: len(t.historyWrapped),
		cursorX:           t.cursorX,
		cursorY:           t.cursorY,
	}
	t.screen = make([]screenLine, len(t.screen))
	t.wrapped = make([]bool, len(t.wrapped))
	t.cursorX, t.cursorY = 0, 0
}

// leaveAlternateScreen restores the main screen and drops the alternate screen
// content, including lines it may have scrolled into the scrollback.
func (t *terminal) leaveAlternateScreen() {
	if t.alternate == nil {
		return
	}
	rows := len(t.screen)
	screen := t.alternate.screen
	if len(screen) > rows {
		screen = screen[:rows]
	}
	for len(screen) < rows {
		screen = append(screen, nil)
	}
	t.screen = screen
	wrapped := t.alternate.wrapped
	if len(wrapped) > rows {
		wrapped = wrapped[:rows]
	}
	for len(wrapped) < rows {
		wrapped = append(wrapped, false)
	}
	t.wrapped = wrapped
	if len(t.history) > t.alternate.historyLen {
		t.history = t.history[:t.alternate.historyLen]
	}
	if len(t.historyWrapped) > t.alternate.historyWrappedLen {
		t.historyWrapped = t.historyWrapped[:t.alternate.historyWrappedLen]
	}
	t.cursorX = clamp(t.alternate.cursorX, 0, t.columns-1)
	t.cursorY = clamp(t.alternate.cursorY, 0, len(t.screen)-1)
	t.alternate = nil
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
	end := min(t.cursorX+count, len(line))
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
	// Build a new line: appending to line[:cursorX] in place would overwrite
	// the characters that are about to be shifted to the right.
	inserted := make(screenLine, 0, len(line)+count)
	inserted = append(inserted, line[:t.cursorX]...)
	inserted = append(inserted, blanks...)
	inserted = append(inserted, line[t.cursorX:]...)
	t.screen[t.cursorY] = inserted
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
	copy(t.wrapped[t.cursorY+count:], t.wrapped[t.cursorY:rows-count])
	for index := 0; index < count; index++ {
		t.screen[t.cursorY+index] = nil
		t.wrapped[t.cursorY+index] = false
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
	copy(t.wrapped[t.cursorY:], t.wrapped[t.cursorY+count:])
	for index := rows - count; index < rows; index++ {
		t.screen[index] = nil
		t.wrapped[index] = false
	}
}

// scrollUp moves count lines of the viewport into the scrollback.
func (t *terminal) scrollUp(count int) {
	for range count {
		t.scroll()
	}
}

// scrollDown moves the viewport content down, dropping the bottom lines.
func (t *terminal) scrollDown(count int) {
	for index := 0; index < count && len(t.screen) > 0; index++ {
		copy(t.screen[1:], t.screen[:len(t.screen)-1])
		t.screen[0] = nil
		copy(t.wrapped[1:], t.wrapped[:len(t.wrapped)-1])
		t.wrapped[0] = false
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
