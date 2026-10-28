package utils

import (
	"sort"

	"github.com/leafevans/go-fuck/internal/conf"
)

// SequenceRatio returns the similarity of two strings, mirroring the
// SequenceMatcher.ratio of the Python standard library. The result is between
// 0 and 1.
func SequenceRatio(a, b string) float64 {
	left, right := []rune(a), []rune(b)
	length := len(left) + len(right)
	if length == 0 {
		return 1
	}
	return 2 * float64(matchingBlockSize(left, right)) / float64(length)
}

// matchingBlockSize returns the total size of the matching blocks of a and b.
func matchingBlockSize(a, b []rune) int {
	type block struct {
		alo, ahi, blo, bhi int
	}
	type match struct {
		i, j, size int
	}

	queue := []block{{0, len(a), 0, len(b)}}
	blocks := []match{}
	for len(queue) > 0 {
		current := queue[len(queue)-1]
		queue = queue[:len(queue)-1]

		i, j, size := findLongestMatch(a, b, current.alo, current.ahi, current.blo, current.bhi)
		if size == 0 {
			continue
		}
		blocks = append(blocks, match{i, j, size})
		if current.alo < i && current.blo < j {
			queue = append(queue, block{current.alo, i, current.blo, j})
		}
		if i+size < current.ahi && j+size < current.bhi {
			queue = append(queue, block{i + size, current.ahi, j + size, current.bhi})
		}
	}

	total := 0
	for _, block := range blocks {
		total += block.size
	}
	return total
}

// findLongestMatch returns the longest matching block inside the given ranges.
func findLongestMatch(a, b []rune, alo, ahi, blo, bhi int) (int, int, int) {
	bestI, bestJ, bestSize := alo, blo, 0
	previous := map[int]int{}
	for i := alo; i < ahi; i++ {
		current := map[int]int{}
		for j := blo; j < bhi; j++ {
			if a[i] != b[j] {
				continue
			}
			size := previous[j-1] + 1
			current[j] = size
			if size > bestSize {
				bestI, bestJ, bestSize = i-size+1, j-size+1, size
			}
		}
		previous = current
	}
	return bestI, bestJ, bestSize
}

// GetCloseMatches returns at most n entries of possibilities that are similar
// to word, ordered by similarity.
//
// A non-positive n falls back to the num_close_matches setting.
func GetCloseMatches(word string, possibilities []string, n int, cutoff float64) []string {
	if n <= 0 {
		n = conf.Current().NumCloseMatches
	}
	if n <= 0 {
		return nil
	}

	type scored struct {
		ratio float64
		value string
	}
	matches := []scored{}
	for _, possible := range possibilities {
		if ratio := SequenceRatio(possible, word); ratio >= cutoff {
			matches = append(matches, scored{ratio, possible})
		}
	}

	sort.SliceStable(matches, func(left, right int) bool {
		return matches[left].ratio > matches[right].ratio
	})

	if len(matches) > n {
		matches = matches[:n]
	}
	result := make([]string, len(matches))
	for i, match := range matches {
		result[i] = match.value
	}
	return result
}

// GetClosest returns the closest match, or the first possibility when the
// fallback is enabled and nothing matches.
func GetClosest(word string, possibilities []string, cutoff float64, fallbackToFirst bool) string {
	if matches := GetCloseMatches(word, possibilities, 1, cutoff); len(matches) > 0 {
		return matches[0]
	}
	if fallbackToFirst && len(possibilities) > 0 {
		return possibilities[0]
	}
	return ""
}

// GetClosestDefault returns the closest match with the default cutoff of 0.6,
// falling back to the first possibility.
func GetClosestDefault(word string, possibilities []string) string {
	return GetClosest(word, possibilities, 0.6, true)
}
