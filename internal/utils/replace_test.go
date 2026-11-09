package utils

import "testing"

func TestReplaceArgumentsSorted(t *testing.T) {
	got := ReplaceArgumentsSorted("uv pip instol", "instol",
		[]string{"uninstall", "list", "install"})
	want := []string{"uv pip install", "uv pip uninstall", "uv pip list"}
	if len(got) != len(want) {
		t.Fatalf("ReplaceArgumentsSorted() = %q, want %q", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("ReplaceArgumentsSorted() = %q, want %q", got, want)
		}
	}
}

func TestFindLineByPrefix(t *testing.T) {
	testCases := []struct {
		name   string
		output string
		prefix string
		want   string
		found  bool
	}{
		{"exact", "a\nb\nc\n", "b", "b", true},
		{"leading whitespace", "\n\tgit branch --set-upstream-to=origin/main main\n\n\n", "git branch --set-upstream-to", "git branch --set-upstream-to=origin/main main", true},
		{"crlf", "a\r\nb\r\n", "b", "b", true},
		{"missing", "a\nb\n", "c", "", false},
		{"empty", "", "a", "", false},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got, found := FindLineByPrefix(testCase.output, testCase.prefix)
			if got != testCase.want || found != testCase.found {
				t.Fatalf("FindLineByPrefix(%q, %q) = %q, %t; want %q, %t",
					testCase.output, testCase.prefix, got, found, testCase.want, testCase.found)
			}
		})
	}
}
