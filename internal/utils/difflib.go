package utils

import (
	"sort"
	"unicode/utf8"

	"github.com/leafevans/go-fuck/internal/conf"
)

// SequenceRatio returns the similarity of two strings, mirroring the
// SequenceMatcher.ratio of the Python standard library. The result is between
// 0 and 1.
func SequenceRatio(a, b string) float64 {
	return ratioRunes([]rune(a), []rune(b))
}

// ratioRunes returns the similarity of two rune slices.
func ratioRunes(left, right []rune) float64 {
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

	queue := []block{{0, len(a), 0, len(b)}}
	// Two rows of the dynamic programming table are reused for every
	// findLongestMatch call, so the hot path does not allocate per row.
	previous := make([]int, len(b)+1)
	current := make([]int, len(b)+1)

	total := 0
	for len(queue) > 0 {
		currentBlock := queue[len(queue)-1]
		queue = queue[:len(queue)-1]

		i, j, size := findLongestMatch(a, b, currentBlock.alo, currentBlock.ahi,
			currentBlock.blo, currentBlock.bhi, previous, current)
		if size == 0 {
			continue
		}
		total += size
		if currentBlock.alo < i && currentBlock.blo < j {
			queue = append(queue, block{currentBlock.alo, i, currentBlock.blo, j})
		}
		if i+size < currentBlock.ahi && j+size < currentBlock.bhi {
			queue = append(queue, block{i + size, currentBlock.ahi, j + size, currentBlock.bhi})
		}
	}
	return total
}

// findLongestMatch returns the longest matching block inside the given ranges.
//
// The previous and current rows of the dynamic programming table are passed in
// by the caller so that they can be reused across calls.
func findLongestMatch(a, b []rune, alo, ahi, blo, bhi int, previous, current []int) (int, int, int) {
	bestI, bestJ, bestSize := alo, blo, 0
	width := bhi - blo
	for cell := 0; cell <= width; cell++ {
		previous[cell] = 0
		current[cell] = 0
	}

	for i := alo; i < ahi; i++ {
		current[0] = 0
		for j := blo; j < bhi; j++ {
			cell := j - blo + 1
			if a[i] != b[j] {
				current[cell] = 0
				continue
			}
			size := previous[cell-1] + 1
			current[cell] = size
			if size > bestSize {
				bestI = i - size + 1
				bestJ = j - size + 1
				bestSize = size
			}
		}
		previous, current = current, previous
	}
	return bestI, bestJ, bestSize
}

// realQuickRatio returns a cheap upper bound of the similarity of two strings,
// mirroring SequenceMatcher.real_quick_ratio.
func realQuickRatio(leftLen, rightLen int) float64 {
	length := leftLen + rightLen
	if length == 0 {
		return 1
	}
	if leftLen < rightLen {
		return 2 * float64(leftLen) / float64(length)
	}
	return 2 * float64(rightLen) / float64(length)
}

// quickRatioUpperBound returns a cheap upper bound of the similarity of a
// candidate and the word, mirroring SequenceMatcher.quick_ratio.
//
// The upper bound never underestimates the real ratio, so candidates below the
// cutoff can be skipped without changing the result.
func quickRatioUpperBound(candidate string, wordCounts map[rune]int, scratch map[rune]int, wordLen, candidateLen int) float64 {
	length := wordLen + candidateLen
	if length == 0 {
		return 1
	}

	for key := range scratch {
		scratch[key] = 0
	}
	for _, char := range candidate {
		if _, ok := wordCounts[char]; ok {
			scratch[char]++
		}
	}

	matching := 0
	for char, count := range wordCounts {
		found := scratch[char]
		if found > count {
			found = count
		}
		matching += found
	}
	return 2 * float64(matching) / float64(length)
}

// runeCounts returns the character frequencies of a rune slice.
func runeCounts(runes []rune) map[rune]int {
	counts := make(map[rune]int, len(runes))
	for _, char := range runes {
		counts[char]++
	}
	return counts
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

	wordRunes := []rune(word)
	wordLen := len(wordRunes)
	var wordCounts map[rune]int
	var scratch map[rune]int

	type scored struct {
		ratio float64
		value string
	}
	matches := make([]scored, 0, n)

	for _, possible := range possibilities {
		possibleLen := utf8.RuneCountInString(possible)
		// Both bounds are upper bounds of the real ratio, so skipping a
		// candidate that cannot reach the cutoff preserves the result.
		if realQuickRatio(wordLen, possibleLen) < cutoff {
			continue
		}
		if wordCounts == nil {
			wordCounts = runeCounts(wordRunes)
			scratch = make(map[rune]int, len(wordCounts))
		}
		if quickRatioUpperBound(possible, wordCounts, scratch, wordLen, possibleLen) < cutoff {
			continue
		}

		if ratio := ratioRunes([]rune(possible), wordRunes); ratio >= cutoff {
			matches = append(matches, scored{ratio, possible})
		}
	}

	// Python uses heapq.nlargest over (ratio, candidate) tuples, so equal
	// ratios are ordered by the candidate string, descending.
	sort.SliceStable(matches, func(left, right int) bool {
		if matches[left].ratio != matches[right].ratio {
			return matches[left].ratio > matches[right].ratio
		}
		return matches[left].value > matches[right].value
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
