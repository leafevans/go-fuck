package system

import (
	"io"
	"os"

	"github.com/leafevans/go-fuck/internal/consts"
)

// keyReader abstracts the raw byte source of the terminal.
type keyReader func() (byte, error)

// stdinByteReader reads single bytes from standard input.
func stdinByteReader() (byte, error) {
	var buf [1]byte
	n, err := os.Stdin.Read(buf[:])
	if n == 1 {
		return buf[0], nil
	}
	if err == nil {
		err = io.EOF
	}
	return 0, err
}

// parseKey converts a raw byte stream into a semantic key.
//
// Escape sequences are mapped to the cursor keys, control characters to the
// corresponding semantic keys and everything else is returned as a one byte
// string key.
func parseKey(read keyReader) (consts.Key, error) {
	ch, err := read()
	if err != nil {
		return "", err
	}

	switch ch {
	case 0x0e: // Ctrl+N
		return consts.KeyCtrlN, nil
	case 0x03: // Ctrl+C
		return consts.KeyCtrlC, nil
	case 0x10: // Ctrl+P
		return consts.KeyCtrlP, nil
	case 0x1b: // Escape sequence
		second, err := read()
		if err != nil {
			return consts.Key(string(rune(ch))), nil
		}
		if second != '[' {
			return consts.Key(string([]byte{ch, second})), nil
		}
		third, err := read()
		if err != nil {
			return consts.Key(string([]byte{ch, second})), nil
		}
		switch third {
		case 'A':
			return consts.KeyUp, nil
		case 'B':
			return consts.KeyDown, nil
		}
		return consts.Key(string([]byte{ch, second, third})), nil
	}

	return consts.Key(string(rune(ch))), nil
}
