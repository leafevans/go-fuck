package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testHash returns a deterministic 64 character checksum.
func testHash(seed byte) string {
	return strings.Repeat(string([]byte{'a' + seed}), 64)
}

// checksumsContent returns a valid checksums.txt for a version.
func checksumsContent(version string) string {
	assets := []struct {
		osName string
		arch   string
		ext    string
		seed   byte
	}{
		{"darwin", "amd64", "tar.gz", 0},
		{"darwin", "arm64", "tar.gz", 1},
		{"linux", "amd64", "tar.gz", 2},
		{"linux", "arm64", "tar.gz", 3},
		{"windows", "amd64", "zip", 4},
		{"windows", "arm64", "zip", 5},
	}
	lines := make([]string, 0, len(assets))
	for _, asset := range assets {
		lines = append(lines, testHash(asset.seed)+"  "+projectName+"_"+version+"_"+asset.osName+"_"+asset.arch+"."+asset.ext)
	}
	return strings.Join(lines, "\n") + "\n"
}

// writeFile writes content to a temporary file and returns its path.
func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

// loadTestRelease loads a valid release for the tests.
func loadTestRelease(t *testing.T) *release {
	t.Helper()
	rel, err := loadRelease(writeFile(t, "checksums.txt", checksumsContent("0.2.0")), "0.2.0")
	if err != nil {
		t.Fatalf("loadRelease() error = %v", err)
	}
	return rel
}

func TestLoadRelease(t *testing.T) {
	rel := loadTestRelease(t)
	if rel.version != "0.2.0" {
		t.Fatalf("version = %q", rel.version)
	}
	if want := testHash(5); rel.hashes["windows_arm64"] != want {
		t.Fatalf("windows_arm64 hash = %q, want %q", rel.hashes["windows_arm64"], want)
	}
	if got, want := rel.asset("linux", "amd64"), "go-fuck_0.2.0_linux_amd64.tar.gz"; got != want {
		t.Fatalf("asset() = %q, want %q", got, want)
	}
	if got, want := rel.url("windows", "arm64"), releaseBase+"/v0.2.0/go-fuck_0.2.0_windows_arm64.zip"; got != want {
		t.Fatalf("url() = %q, want %q", got, want)
	}
}

func TestLoadReleaseRejectsWrongVersion(t *testing.T) {
	path := writeFile(t, "checksums.txt", checksumsContent("0.1.0"))
	if _, err := loadRelease(path, "0.2.0"); err == nil {
		t.Fatal("loadRelease() should reject checksums of another version")
	}
}

func TestLoadReleaseRejectsMissingAsset(t *testing.T) {
	content := strings.Replace(checksumsContent("0.2.0"), testHash(5)+"  "+projectName+"_0.2.0_windows_arm64.zip\n", "", 1)
	path := writeFile(t, "checksums.txt", content)
	if _, err := loadRelease(path, "0.2.0"); err == nil {
		t.Fatal("loadRelease() should reject an incomplete checksums file")
	}
}

func TestFormula(t *testing.T) {
	rel := loadTestRelease(t)
	formulaText := formula(rel)
	for _, want := range []string{
		"version \"0.2.0\"",
		"go-fuck_0.2.0_darwin_arm64.tar.gz",
		testHash(0),
		testHash(3),
	} {
		if !strings.Contains(formulaText, want) {
			t.Fatalf("formula does not contain %q:\n%s", want, formulaText)
		}
	}
}

func TestScoopManifest(t *testing.T) {
	rel := loadTestRelease(t)
	data, err := scoopManifest(rel)
	if err != nil {
		t.Fatalf("scoopManifest() error = %v", err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("scoop manifest is not valid JSON: %v", err)
	}
	if manifest["version"] != "0.2.0" {
		t.Fatalf("version = %v", manifest["version"])
	}
	architectures := manifest["architecture"].(map[string]any)
	amd64 := architectures["64bit"].(map[string]any)
	if amd64["hash"] != testHash(4) {
		t.Fatalf("64bit hash = %v", amd64["hash"])
	}
}

func TestAURPKGBUILD(t *testing.T) {
	rel := loadTestRelease(t)
	checksum := testHash(6)
	pkgbuild := aurPKGBUILD(rel, checksum)
	for _, want := range []string{
		"pkgver=0.2.0\n",
		"pkgrel=1\n",
		"arch=('x86_64')",
		"sha256sums=('" + checksum + "')",
		"archive/refs/tags/v$pkgver.tar.gz",
	} {
		if !strings.Contains(pkgbuild, want) {
			t.Fatalf("AUR PKGBUILD does not contain %q:\n%s", want, pkgbuild)
		}
	}
	if strings.Contains(pkgbuild, "aarch64") {
		t.Fatalf("AUR PKGBUILD must target x86_64 only:\n%s", pkgbuild)
	}
}

func TestAURSrcinfo(t *testing.T) {
	rel := loadTestRelease(t)
	checksum := testHash(6)
	srcinfo := aurSrcinfo(rel, checksum)
	for _, want := range []string{
		"pkgver = 0.2.0\n",
		"pkgrel = 1\n",
		"arch = x86_64\n",
		"source = go-fuck-0.2.0.tar.gz::" + archiveBase + "/v0.2.0.tar.gz",
		"sha256sums = " + checksum,
		"pkgname = go-fuck\n",
	} {
		if !strings.Contains(srcinfo, want) {
			t.Fatalf("AUR .SRCINFO does not contain %q:\n%s", want, srcinfo)
		}
	}
}

func TestFileChecksum(t *testing.T) {
	path := writeFile(t, "archive.tar.gz", "hello")
	want := fmt.Sprintf("%x", sha256.Sum256([]byte("hello")))
	got, err := fileChecksum(path)
	if err != nil {
		t.Fatalf("fileChecksum() error = %v", err)
	}
	if got != want {
		t.Fatalf("fileChecksum() = %q, want %q", got, want)
	}
	if _, err := fileChecksum(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("fileChecksum() should fail for a missing file")
	}
}
