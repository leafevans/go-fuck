package utils

import (
	"cmp"
	"math"
	"slices"
	"strings"
	"testing"
)

func TestGetCloseMatchesAgreesWithFullSort(t *testing.T) {
	long := strings.Repeat("\u03bbx", 50)
	candidates := []string{"", "bat", "cat", "cat", "hat", "rat", "mat", "sat", "at", "fat", "art", "act", "tap", "\u03bb", "\ufffd", "\xff", "\u03bb\xff", long, long + "z"}
	for _, word := range []string{"at", "", "\u03bb\ufffd", long} {
		for _, cutoff := range []float64{0, 0.1, 0.6, 1, math.NaN()} {
			var all []closeMatch
			for _, candidate := range candidates {
				if ratio := SequenceRatio(candidate, word); ratio >= cutoff {
					all = append(all, closeMatch{ratio, candidate})
				}
			}
			slices.SortFunc(all, func(a, b closeMatch) int {
				return cmp.Or(cmp.Compare(b.ratio, a.ratio), cmp.Compare(b.value, a.value))
			})
			for _, limit := range []int{1, 2, 5, len(candidates), math.MaxInt} {
				want := make([]string, min(limit, len(all)))
				for i := range want {
					want[i] = all[i].value
				}
				if got := GetCloseMatches(word, candidates, limit, cutoff); !slices.Equal(got, want) {
					t.Fatalf("GetCloseMatches(%q, limit=%d, cutoff=%g) = %q, want %q", word, limit, cutoff, got, want)
				}
			}
		}
	}
}

func TestGetCloseMatchesKeepsLaterTies(t *testing.T) {
	candidates := []string{"bat", "cat", "cat", "hat", "rat", "aaa", "sat", "aat"}
	want := []string{"sat", "rat"}
	if got := GetCloseMatches("at", candidates, 2, 2.0/3); !slices.Equal(got, want) {
		t.Fatalf("GetCloseMatches() = %q, want %q", got, want)
	}
	if got := GetUniqueClosest("at", candidates, 2.0/3); got != "" {
		t.Fatalf("GetUniqueClosest() = %q, want no unique match", got)
	}
}
