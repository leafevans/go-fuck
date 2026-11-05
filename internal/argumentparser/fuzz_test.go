package argumentparser

import (
	"reflect"
	"slices"
	"strings"
	"testing"
)

// fuzzArgvSeparator separates the argv elements of a fuzz input. A NUL byte
// never appears in a real command line, so seeds cannot collide with it.
const fuzzArgvSeparator = "\x00"

// fuzzSeed encodes argv as one fuzz input.
func fuzzSeed(argv ...string) string {
	return strings.Join(argv, fuzzArgvSeparator)
}

// parseFuzzSeeds is shared by FuzzParse and TestParseFuzzSeedCorpus. FuzzParse
// registers it as the initial corpus and the test asserts that every seed
// still parses successfully or fails as expected.
var parseFuzzSeeds = []struct {
	argv    []string
	wantErr bool
}{
	{[]string{"go-fuck"}, false},
	{[]string{"go-fuck", "git", "brnch"}, false},
	{[]string{"go-fuck", "GOFUCK_ARGUMENT_PLACEHOLDER", "-y", "-d"}, false},
	{[]string{"go-fuck", "THEFUCK_ARGUMENT_PLACEHOLDER", "ls"}, false},
	{[]string{"go-fuck", "-v", "-d", "-y", "--enable-experimental-instant-mode"}, false},
	{[]string{"go-fuck", "-h"}, false},
	{[]string{"go-fuck", "--version"}, false},
	{[]string{"go-fuck", "--setup"}, false},
	{[]string{"go-fuck", "-r"}, false},
	{[]string{"go-fuck", "-y", "-r"}, true},
	{[]string{"go-fuck", "-y", "--hard", "-r"}, true},
	{[]string{"go-fuck", "-a"}, false},
	{[]string{"go-fuck", "-a", "my-fuck"}, false},
	{[]string{"go-fuck", "--alias=other"}, false},
	{[]string{"go-fuck", "--alias="}, false},
	{[]string{"go-fuck", "-a", "--"}, false},
	{[]string{"go-fuck", "-a", "-l"}, true},
	{[]string{"go-fuck", "-l", "/tmp/log", "--force-command", "git brnch"}, false},
	{[]string{"go-fuck", "--shell-logger=log", "--force-command=ls"}, false},
	{[]string{"go-fuck", "-l"}, true},
	{[]string{"go-fuck", "-l", ""}, false},
	{[]string{"go-fuck", "-l", "-d"}, false},
	{[]string{"go-fuck", "--shell-logger"}, true},
	{[]string{"go-fuck", "--force-command"}, true},
	{[]string{"go-fuck", "--force-command="}, false},
	{[]string{"go-fuck", "--force-command", "--json"}, false},
	{[]string{"go-fuck", "--unknown"}, true},
	{[]string{"go-fuck", "--", "-d", "ls"}, false},
	{[]string{"go-fuck", "--", ""}, false},
	{[]string{"go-fuck", "--", "--", "-d"}, false},
	{[]string{"go-fuck", "", "-h"}, false},
	{[]string{"", "-d"}, false},
	{[]string{"GOFUCK_ARGUMENT_PLACEHOLDER", "-y"}, false},
	{[]string{"go-fuck", "git", "status", "GOFUCK_ARGUMENT_PLACEHOLDER", "-y"}, false},
	{[]string{"go-fuck", "THEFUCK_ARGUMENT_PLACEHOLDER"}, false},
	{[]string{"go-fuck", "ünïcode", "☃"}, false},
}

// FuzzParse checks that Parse never panics for arbitrary argv, does not modify
// the argument slice it was given and is deterministic.
func FuzzParse(f *testing.F) {
	for _, seed := range parseFuzzSeeds {
		f.Add(fuzzSeed(seed.argv...))
	}
	f.Fuzz(func(t *testing.T, raw string) {
		argv := strings.Split(raw, fuzzArgvSeparator)
		input := slices.Clone(argv)

		args, err := Parse(argv)
		if !slices.Equal(argv, input) {
			t.Fatalf("Parse(%q) modified its argument slice to %q", input, argv)
		}

		again, againErr := Parse(slices.Clone(input))
		if !reflect.DeepEqual(args, again) {
			t.Fatalf("Parse(%q) is not deterministic: %+v != %+v", input, args, again)
		}
		if fuzzErrorString(err) != fuzzErrorString(againErr) {
			t.Fatalf("Parse(%q) is not deterministic: errors %v != %v", input, err, againErr)
		}
		if err != nil {
			return
		}
		if args == nil {
			t.Fatalf("Parse(%q) returned a nil result without an error", input)
		}
		if args.Yes && args.Repeat {
			t.Fatalf("Parse(%q) accepted -y together with -r: %+v", input, args)
		}
	})
}

// fuzzErrorString returns the error message of err, or an empty string for nil.
func fuzzErrorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// TestParseFuzzSeedCorpus keeps the seed corpus of FuzzParse meaningful.
func TestParseFuzzSeedCorpus(t *testing.T) {
	for _, seed := range parseFuzzSeeds {
		args, err := Parse(slices.Clone(seed.argv))
		if (err != nil) != seed.wantErr {
			t.Errorf("Parse(%q) error = %v, wantErr %t", seed.argv, err, seed.wantErr)
			continue
		}
		if err == nil && args == nil {
			t.Errorf("Parse(%q) returned a nil result without an error", seed.argv)
		}
	}
}
