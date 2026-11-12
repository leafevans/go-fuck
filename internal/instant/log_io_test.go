package instant

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/consts"
)

func TestReadLogKeepsNewestBytes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "instant.log")
	content := strings.Repeat("a", consts.LogSizeInBytes) + "tail"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := readLog(path)
	if err != nil {
		t.Fatalf("readLog() error = %v", err)
	}
	if len(got) != consts.LogSizeInBytes {
		t.Fatalf("readLog() length = %d, want %d", len(got), consts.LogSizeInBytes)
	}
	if !strings.HasSuffix(got, "tail") {
		t.Fatalf("readLog() does not keep the newest bytes")
	}
}

func TestReadLogMissingFile(t *testing.T) {
	if _, err := readLog(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("readLog() of a missing file returned no error")
	}
}

func TestTrimFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "log")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if _, err := file.WriteString("abcdef"); err != nil {
		t.Fatal(err)
	}
	if err := trimFile(file, 3); err != nil {
		t.Fatalf("trimFile() error = %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "def" {
		t.Fatalf("trimFile() left %q, want def", string(got))
	}
	if err := trimFile(file, 0); err != nil {
		t.Fatalf("trimFile(0) error = %v", err)
	}
	got, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("trimFile(0) left %q, want empty", string(got))
	}
}
