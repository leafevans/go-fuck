package utils

import (
	"cmp"
	"slices"
	"strings"
	"unicode/utf8"
)

// SequenceRatio returns the similarity of two strings, mirroring the
// SequenceMatcher.ratio of the Python standard library. The result is between
// 0 and 1.
func SequenceRatio(a, b string) float64 {
	if a == b {
		return 1
	}
	shorter, longer := a, b
	if len(a) > len(b) {
		shorter, longer = b, a
	}
	// A complete substring achieves the maximum possible match. Require valid
	// UTF-8 because byte containment is not equivalent to rune containment
	// for malformed strings.
	if len(longer) >= 64 && strings.Contains(longer, shorter) && utf8.ValidString(shorter) && utf8.ValidString(longer) {
		return realQuickRatio(utf8.RuneCountInString(a), utf8.RuneCountInString(b))
	}
	var matcher sequenceMatcher
	return matcher.ratio([]rune(a), []rune(b))
}

// SortBySimilarity returns a copy of candidates sorted by decreasing
// similarity to word. Candidates with equal similarity keep their original
// order. It is used to rank the suggestions that a tool prints before they are
// offered to the user.
func SortBySimilarity(word string, candidates []string) []string {
	type scored struct {
		value string
		ratio float64
	}
	items := make([]scored, len(candidates))
	var matcher sequenceMatcher
	wordRunes := []rune(word)
	for index, candidate := range candidates {
		items[index] = scored{value: candidate, ratio: matcher.ratio(wordRunes, []rune(candidate))}
	}
	slices.SortStableFunc(items, func(left, right scored) int {
		return cmp.Compare(right.ratio, left.ratio)
	})
	sorted := make([]string, len(items))
	for index, item := range items {
		sorted[index] = item.value
	}
	return sorted
}

// sequenceMatcher reuses its dynamic programming rows across comparisons.
type sequenceMatcher struct {
	inline [128]int
	large  []int
}

// ratio returns the similarity score between left and right in [0, 1].
func (m *sequenceMatcher) ratio(left, right []rune) float64 {
	length := len(left) + len(right)
	if length == 0 {
		return 1
	}
	if len(left) == 0 || len(right) == 0 {
		return 0
	}
	return 2 * float64(m.matchingBlockSize(left, right)) / float64(length)
}

// matchingBlockSize returns the total size of the matching blocks of a and b.
func (m *sequenceMatcher) matchingBlockSize(a, b []rune) int {
	type block struct {
		alo, ahi, blo, bhi int
	}

	var pending [8]block
	queue := append(pending[:0], block{0, len(a), 0, len(b)})
	width := len(b) + 1
	rows := m.inline[:]
	if 2*width > len(rows) {
		m.large = slices.Grow(m.large[:0], 2*width)
		rows = m.large[:2*width]
	}
	previous, current := rows[:width], rows[width:2*width]

	total := 0
	for len(queue) > 0 {
		currentBlock := queue[len(queue)-1]
		queue = queue[:len(queue)-1]

		i, j, size := findLongestMatch(a[currentBlock.alo:currentBlock.ahi],
			b[currentBlock.blo:currentBlock.bhi], previous, current)
		if size == 0 {
			continue
		}
		i += currentBlock.alo
		j += currentBlock.blo
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

// findLongestMatch breaks ties by the earliest position in a, then in b.
func findLongestMatch(a, b []rune, previous, current []int) (int, int, int) {
	previous = previous[:len(b)+1]
	current = current[:len(b)+1]
	clear(previous)
	current[0] = 0
	bestI, bestJ, bestSize := 0, 0, 0
	maxSize := min(len(a), len(b))
	for i, left := range a {
		for j, right := range b {
			if left != right {
				current[j+1] = 0
				continue
			}
			size := previous[j] + 1
			current[j+1] = size
			if size > bestSize {
				bestI = i - size + 1
				bestJ = j - size + 1
				bestSize = size
				if size == maxSize {
					return bestI, bestJ, bestSize
				}
			}
		}
		previous, current = current, previous
	}
	return bestI, bestJ, bestSize
}
