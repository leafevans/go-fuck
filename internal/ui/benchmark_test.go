package ui

import (
	"strconv"
	"testing"

	"github.com/leafevans/go-fuck/v4/internal/types"
)

func BenchmarkFzfCandidates(b *testing.B) {
	for _, size := range []int{10, 100, 1000} {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			commands := make([]*types.CorrectedCommand, size)
			for i := range commands {
				commands[i] = &types.CorrectedCommand{Script: "command " + strconv.Itoa(i)}
			}
			b.ReportAllocs()
			for b.Loop() {
				fzfCandidates(commands)
			}
		})
	}
}
