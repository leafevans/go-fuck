package argumentparser

import (
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/consts"
)

func TestParseDoctorAndListRules(t *testing.T) {
	consts.SetExecutableName(consts.AppName)
	for _, testCase := range []struct {
		name      string
		argv      []string
		doctor    bool
		listRules bool
		json      bool
	}{
		{"doctor", []string{"go-fuck", "--doctor"}, true, false, false},
		{"list rules", []string{"go-fuck", "--list-rules"}, false, true, false},
		{"doctor json", []string{"go-fuck", "--doctor", "--json"}, true, false, true},
		{"list rules json", []string{"go-fuck", "--list-rules", "--json"}, false, true, true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			args, err := Parse(testCase.argv)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", testCase.argv, err)
			}
			if args.Doctor != testCase.doctor {
				t.Fatalf("args.Doctor = %t, want %t", args.Doctor, testCase.doctor)
			}
			if args.ListRules != testCase.listRules {
				t.Fatalf("args.ListRules = %t, want %t", args.ListRules, testCase.listRules)
			}
			if args.JSON != testCase.json {
				t.Fatalf("args.JSON = %t, want %t", args.JSON, testCase.json)
			}
		})
	}
}

func TestUsageAndHelpMentionNewFlags(t *testing.T) {
	consts.SetExecutableName(consts.AppName)
	usage := Usage()
	if !strings.Contains(usage, "--doctor") || !strings.Contains(usage, "--list-rules") {
		t.Fatalf("Usage() = %q, does not mention the new flags", usage)
	}
	help := Help()
	for _, want := range []string{"--doctor", "diagnostic report", "--list-rules", "built-in rules"} {
		if !strings.Contains(help, want) {
			t.Fatalf("Help() does not contain %q:\n%s", want, help)
		}
	}
}

func TestParseOutputFrom(t *testing.T) {
	consts.SetExecutableName(consts.AppName)

	args, err := Parse([]string{"go-fuck", "--force-command", "git push", "--output-from", "failed.log"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if args.ForceCommand != "git push" || args.OutputFrom != "failed.log" {
		t.Fatalf("unexpected args: %+v", args)
	}

	args, err = Parse([]string{"go-fuck", "--force-command=git push", "--output-from=-"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if args.OutputFrom != "-" {
		t.Fatalf("args.OutputFrom = %q, want -", args.OutputFrom)
	}

	if _, err := Parse([]string{"go-fuck", "--force-command", "git push", "--output-from"}); err == nil {
		t.Fatal("Parse() accepted --output-from without a value")
	}
	if _, err := Parse([]string{"go-fuck", "--output-from", "failed.log"}); err == nil {
		t.Fatal("Parse() accepted --output-from without --force-command")
	}

	help := Help()
	for _, want := range []string{"--output-from FILE", "instead of re-running it"} {
		if !strings.Contains(help, want) {
			t.Fatalf("Help() does not contain %q:\n%s", want, help)
		}
	}
	if !strings.Contains(Usage(), "--output-from") {
		t.Fatalf("Usage() = %q, does not mention --output-from", Usage())
	}
}
