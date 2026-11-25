package consts

import "testing"

func TestBuildModule(t *testing.T) {
	path, version := BuildModule()
	if path != "github.com/leafevans/go-fuck/v4" {
		t.Fatalf("BuildModule() path = %q, want the v2 module path", path)
	}
	if version == "" {
		t.Fatal("BuildModule() version is empty")
	}
}
