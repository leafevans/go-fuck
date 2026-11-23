package utils

import (
	"cmp"
	"slices"
	"unicode/utf8"

	"github.com/leafevans/go-fuck/v3/internal/conf"
)

// realQuickRatio returns a cheap upper bound of the similarity of two strings,
// mirroring SequenceMatcher.real_quick_ratio.
func realQuickRatio(leftLen, rightLen int) float64 {
	length := leftLen + rightLen
	if length == 0 {
		return 1
	}
	return 2 * float64(min(leftLen, rightLen)) / float64(length)
}

// quickRatioUpperBound counts shared characters without regard to order,
// mirroring SequenceMatcher.quick_ratio.
//
// The upper bound never underestimates the real ratio, so candidates below the
// cutoff can be skipped without changing the result.
func (counts *closeMatchCounts) quickRatioUpperBound(candidate string, wordLen, candidateLen int) float64 {
	length := wordLen + candidateLen
	if length == 0 {
		return 1
	}

	var asciiUsed [utf8.RuneSelf]int
	clear(counts.otherUsed)
	matching := 0
	for _, char := range candidate {
		if uint32(char) < utf8.RuneSelf {
			if asciiUsed[char] < counts.ascii[char] {
				asciiUsed[char]++
				matching++
			}
		} else if limit := counts.other[char]; limit != 0 && counts.otherUsed[char] < limit {
			counts.otherUsed[char]++
			matching++
		}
	}
	return 2 * float64(matching) / float64(length)
}

type closeMatchCounts struct {
	// Command names usually use ASCII. Only allocate maps for other runes.
	ascii     [utf8.RuneSelf]int
	other     map[rune]int
	otherUsed map[rune]int
}

func (counts *closeMatchCounts) init(runes []rune) {
	for _, char := range runes {
		if uint32(char) < utf8.RuneSelf {
			counts.ascii[char]++
		} else {
			if counts.other == nil {
				counts.other = make(map[rune]int)
				counts.otherUsed = make(map[rune]int)
			}
			counts.other[char]++
		}
	}
}

type closeMatch struct {
	ratio float64
	value string
}

// Python ranks (ratio, candidate) tuples in descending order.
func compareCloseMatches(left, right closeMatch) int {
	if order := cmp.Compare(right.ratio, left.ratio); order != 0 {
		return order
	}
	return cmp.Compare(right.value, left.value)
}

// canPlace also accepts an upper bound. A tie can enter a full shortlist
// only if the candidate sorts before its last entry.
func canPlace(ratio float64, value string, threshold closeMatch, pruned bool) bool {
	return ratio > threshold.ratio ||
		(ratio == threshold.ratio && (!pruned || value > threshold.value))
}

// GetCloseMatches returns at most n entries of possibilities that are similar
// to word, ordered by similarity.
//
// A non-positive n falls back to the num_close_matches setting.
func GetCloseMatches(word string, possibilities []string, n int, cutoff float64) []string {
	var inline [8]closeMatch
	matches := rankedCloseMatches(inline[:0], word, possibilities, n, cutoff)
	if matches == nil {
		return nil
	}
	result := make([]string, len(matches))
	for i, match := range matches {
		result[i] = match.value
	}
	return result
}

// rankedCloseMatches uses the caller's empty scratch slice for the shortlist.
func rankedCloseMatches(matches []closeMatch, word string, possibilities []string, n int, cutoff float64) []closeMatch {
	if n <= 0 {
		n = conf.Current().NumCloseMatches
	}
	if n <= 0 {
		return nil
	}
	n = min(n, len(possibilities))
	if n == 0 {
		return matches
	}
	if !(cutoff <= 1) { // NaN also rejects every score.
		return matches
	}

	wordRunes := []rune(word)
	var matcher sequenceMatcher
	wordLen := len(wordRunes)
	var wordCounts closeMatchCounts
	countsReady := false

	// Keep at most one extra batch before discarding candidates outside the top n.
	batchLimit := n + min(n, len(possibilities)-n)
	threshold := closeMatch{ratio: cutoff}
	pruned := false

	for _, possible := range possibilities {
		ratio := 1.0
		if possible != word {
			if threshold.ratio > 0 || pruned {
				possibleLen := utf8.RuneCountInString(possible)
				if !canPlace(realQuickRatio(wordLen, possibleLen), possible, threshold, pruned) {
					continue
				}
				if !countsReady {
					wordCounts.init(wordRunes)
					countsReady = true
				}
				upper := wordCounts.quickRatioUpperBound(possible, wordLen, possibleLen)
				if !canPlace(upper, possible, threshold, pruned) {
					continue
				}
			}
			ratio = matcher.ratio([]rune(possible), wordRunes)
		}

		if canPlace(ratio, possible, threshold, pruned) {
			matches = append(matches, closeMatch{ratio, possible})
			if len(matches) == batchLimit && batchLimit < len(possibilities) {
				slices.SortFunc(matches, compareCloseMatches)
				clear(matches[n:])
				matches = matches[:n]
				threshold = matches[n-1]
				pruned = true
			}
		}
	}

	slices.SortFunc(matches, compareCloseMatches)

	if len(matches) > n {
		matches = matches[:n]
	}
	return matches
}

// GetUniqueClosest returns the best match only when no other candidate ties it.
func GetUniqueClosest(word string, possibilities []string, cutoff float64) string {
	var inline [4]closeMatch
	matches := rankedCloseMatches(inline[:0], word, possibilities, 2, cutoff)
	if len(matches) == 0 || (len(matches) == 2 && matches[0].ratio == matches[1].ratio) {
		return ""
	}
	return matches[0].value
}

// GetClosest returns the closest match, or the first possibility when the
// fallback is enabled and nothing matches.
func GetClosest(word string, possibilities []string, cutoff float64, fallbackToFirst bool) string {
	var inline [2]closeMatch
	if matches := rankedCloseMatches(inline[:0], word, possibilities, 1, cutoff); len(matches) > 0 {
		return matches[0].value
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
