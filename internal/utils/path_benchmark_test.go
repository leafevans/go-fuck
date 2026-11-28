package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v4/internal/system"
)

func BenchmarkPathLookup(b *testing.B) {
	root := b.TempDir()
	dirs := make([]string, 32)
	for i := range dirs {
		dirs[i] = filepath.Join(root, fmt.Sprint(i))
		if err := os.Mkdir(dirs[i], 0700); err != nil {
			b.Fatal(err)
		}
	}
	const found = "benchmark-path-tool.exe"
	const missing = "benchmark-path-absent.exe"
	if err := os.WriteFile(filepath.Join(dirs[len(dirs)-1], found), nil, 0700); err != nil {
		b.Fatal(err)
	}
	path := strings.Join(dirs, string(os.PathListSeparator))
	b.Setenv("PATH", path)
	if system.Which(found) == "" || system.Which(missing) != "" {
		b.Fatal("unexpected lookup result")
	}
	b.Run("directory_scan", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			readPathExecutableNames(path, "", nil)
		}
	})
	for _, tc := range []struct {
		name, program string
		lookup        func(string) string
	}{
		{"uncached_late_hit", found, system.Which}, {"uncached_miss", missing, system.Which},
		{"cached_hit", found, Which}, {"cached_miss", missing, Which},
	} {
		b.Run(tc.name, func(b *testing.B) {
			tc.lookup(tc.program)
			b.ReportAllocs()
			for b.Loop() {
				tc.lookup(tc.program)
			}
		})
	}
}
