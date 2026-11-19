package utils

import "testing"

func TestSortBySimilarity(t *testing.T) {
	got := SortBySimilarity("alpha", []string{"beta", "alpha", "alphi"})
	if len(got) != 3 || got[0] != "alpha" || got[1] != "alphi" {
		t.Fatalf("SortBySimilarity() = %q, want the closest candidates first", got)
	}
}

func TestSortBySimilarityKeepsTies(t *testing.T) {
	got := SortBySimilarity("zzz", []string{"first", "second", "third"})
	want := []string{"first", "second", "third"}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("SortBySimilarity() = %q, want stable order %q", got, want)
		}
	}
}

// TestSortBySimilarityAllocations guards the ranking path against a comparator
// that computes the similarity again: recomputing it inside the sort allocates
// a map and rune slices for every comparison, which shows up as a multiple of
// the one-time scoring of the candidates.
func TestSortBySimilarityAllocations(t *testing.T) {
	candidates := buildExecutableNames(200)
	allocs := testing.AllocsPerRun(50, func() {
		SortBySimilarity("gti", candidates)
	})
	// The current implementation needs about 2 allocations per candidate; the
	// bound leaves room for runtime changes while catching a comparator that
	// scores the candidates repeatedly.
	if allocs > 1000 {
		t.Fatalf("SortBySimilarity allocated %.0f objects per call, want at most 1000", allocs)
	}
}
