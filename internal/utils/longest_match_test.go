package utils

import (
	"strings"
	"testing"
)

func TestFindLongestMatchExhaustive(t *testing.T) {
	inputs := [][]rune{nil}
	for size := 1; size <= 5; size++ {
		for bits := range 1 << size {
			word := make([]rune, size)
			for i := range word {
				word[i] = 'a' + rune(bits>>i&1)
			}
			inputs = append(inputs, word)
		}
	}

	// Reuse dirty scratch space as the production caller does.
	var previous, current [6]int
	for _, a := range inputs {
		for _, b := range inputs {
			wantI, wantJ, wantSize := 0, 0, 0
			for i := range a {
				for j := range b {
					size := 0
					for i+size < len(a) && j+size < len(b) && a[i+size] == b[j+size] {
						size++
					}
					if size > wantSize {
						wantI, wantJ, wantSize = i, j, size
					}
				}
			}
			i, j, size := findLongestMatch(a, b, previous[:], current[:])
			if i != wantI || j != wantJ || size != wantSize {
				t.Fatalf("findLongestMatch(%q, %q) = (%d, %d, %d), want (%d, %d, %d)",
					string(a), string(b), i, j, size, wantI, wantJ, wantSize)
			}
		}
	}
}

func TestSequenceRatioIdentical(t *testing.T) {
	if got := SequenceRatio("git status", "git status"); got != 1 {
		t.Fatalf("SequenceRatio of identical strings = %g, want 1", got)
	}
}

func TestSequenceRatioLongContainment(t *testing.T) {
	path := strings.Repeat("x", 80)
	for _, pair := range []struct {
		name, left, right string
		want              float64
	}{
		{"ascii", "add", path + "add", 6.0 / 86},
		{"unicode", "\u76ee\u5f55", path + "\u76ee\u5f55", 4.0 / 84},
		{"invalidByte", "\xa9", path + "\u00e9", 0},
		{"replacementRune", "\ufffd", path + "\xff", 2.0 / 82},
	} {
		t.Run(pair.name, func(t *testing.T) {
			if got := SequenceRatio(pair.left, pair.right); got != pair.want {
				t.Fatalf("forward score = %g, want %g", got, pair.want)
			}
			if got := SequenceRatio(pair.right, pair.left); got != pair.want {
				t.Fatalf("reverse score = %g, want %g", got, pair.want)
			}
		})
	}
}
