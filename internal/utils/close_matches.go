package utils

import (
	"cmp"
	"slices"
	"unicode/utf8"

	"github.com/leafevans/go-fuck/v2/internal/conf"
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

	clear(scratch)
	matching := 0
	for _, char := range candidate {
		if limit := wordCounts[char]; limit != 0 {
			if count := scratch[char]; count < limit {
				scratch[char] = count + 1
				matching++
			}
		}
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

// GetCloseMatches returns at most n entries of possibilities that are similar
// to word, ordered by similarity.
//
// A non-positive n falls back to the num_close_matches setting.
func GetCloseMatches(word string, possibilities []string, n int, cutoff float64) []string {
	matches := rankedCloseMatches(word, possibilities, n, cutoff)
	if matches == nil {
		return nil
	}
	result := make([]string, len(matches))
	for i, match := range matches {
		result[i] = match.value
	}
	return result
}

func rankedCloseMatches(word string, possibilities []string, n int, cutoff float64) []closeMatch {
	if n <= 0 {
		n = conf.Current().NumCloseMatches
	}
	if n <= 0 {
		return nil
	}
	n = min(n, len(possibilities))
	if n == 0 {
		return []closeMatch{}
	}

	wordRunes := []rune(word)
	var matcher sequenceMatcher
	wordLen := len(wordRunes)
	var wordCounts map[rune]int
	var scratch map[rune]int

	// Keep at most one extra batch before discarding candidates outside the top n.
	batchLimit := n + min(n, len(possibilities)-n)
	var inline [8]closeMatch
	matches := inline[:0]
	var worst closeMatch
	pruned := false

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

		if ratio := matcher.ratio([]rune(possible), wordRunes); ratio >= cutoff {
			match := closeMatch{ratio, possible}
			if pruned && compareCloseMatches(match, worst) >= 0 {
				continue
			}
			matches = append(matches, match)
			if len(matches) == batchLimit && batchLimit < len(possibilities) {
				slices.SortFunc(matches, compareCloseMatches)
				clear(matches[n:])
				matches = matches[:n]
				worst = matches[n-1]
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
	matches := rankedCloseMatches(word, possibilities, 2, cutoff)
	if len(matches) == 0 || (len(matches) == 2 && matches[0].ratio == matches[1].ratio) {
		return ""
	}
	return matches[0].value
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
