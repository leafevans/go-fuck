//go:build windows

package system

import "github.com/leafevans/go-fuck/internal/consts"

// GetKey waits for and returns the next key press.
func GetKey() (consts.Key, error) {
	initInput()
	return parseKey(stdinByteReader)
}
