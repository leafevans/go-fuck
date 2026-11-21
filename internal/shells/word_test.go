package shells

import "testing"

func TestReadWord(t *testing.T) {
	for _, tc := range []struct {
		name, raw, rest, word string
		shell                 Shell
	}{
		{"plain", "alpha", "", "alpha", NewGeneric()},
		{"separator", "alpha", "&next", "alpha", NewBash()},
		{"quoted", `"two words"`, "  tail", "two words", NewZsh()},
		{"quoted-empty", `''`, " tail", "", NewBash()},
		{"escaped-posix", `x\ y`, "\ttail", `x\ y`, NewBash()},
		{"escaped-fish", `'x\' y'`, " tail", "x' y", NewFish()},
		{"escaped-powershell", "a` b", " tail", "a b", NewPowershell()},
		{"quoted-powershell", `'a'' b'`, " tail", "a' b", NewPowershell()},
		{"windows-path", `C:\work\file`, " tail", `C:\work\file`, NewPowershell()},
		{"unterminated", "", `word"unfinished`, "", NewBash()},
		{"operator", "", "| next", "", NewGeneric()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if word, size := ReadWord(tc.raw+tc.rest, tc.shell); word != tc.word || size != len(tc.raw) {
				t.Fatalf("ReadWord(%q) = %q, %d; want %q, %d", tc.raw+tc.rest, word, size, tc.word, len(tc.raw))
			}
		})
	}
}
