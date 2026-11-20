package utils

import "testing"

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
