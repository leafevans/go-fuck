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
