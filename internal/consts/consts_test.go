package consts

import "testing"

func TestExecutableName(t *testing.T) {
	t.Cleanup(func() { SetExecutableName(AppName) })

	cases := []struct {
		raw  string
		want string
	}{
		{AppName, AppName},
		{"gof", "gof"},
		{"gof.exe", "gof"},
		{"GOF.EXE", "GOF"},
		{"C:/Program Files/go-fuck/gof.exe", "gof"},
		{"/usr/local/bin/go-fuck", "go-fuck"},
		{"", AppName},
	}
	for _, testCase := range cases {
		SetExecutableName(testCase.raw)
		if got := ExecutableName(); got != testCase.want {
			t.Errorf("SetExecutableName(%q): ExecutableName() = %q, want %q", testCase.raw, got, testCase.want)
		}
	}
}
