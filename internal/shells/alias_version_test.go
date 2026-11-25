package shells

import (
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v4/internal/consts"
)

func TestGeneratedAliasesUseCurrentFormatVersion(t *testing.T) {
	for _, shell := range []Shell{NewBash(), NewZsh(), NewFish(), NewTcsh(), NewPowershell()} {
		t.Run(shell.Name(), func(t *testing.T) {
			alias := shell.AppAlias("fuck")
			if !strings.Contains(alias, consts.AliasVersion) {
				t.Fatalf("%s alias does not contain format version %s:\n%s",
					shell.Name(), consts.AliasVersion, alias)
			}
		})
	}
}

func TestPowershellAliasSkipsConsecutiveAliasHistory(t *testing.T) {
	alias := NewPowershell().AppAlias("fuck")
	for _, want := range []string{"Get-History -Count 8", "-ne \"fuck\"", "Select-Object -Last 1"} {
		if !strings.Contains(alias, want) {
			t.Fatalf("PowerShell alias does not contain %q:\n%s", want, alias)
		}
	}
}
