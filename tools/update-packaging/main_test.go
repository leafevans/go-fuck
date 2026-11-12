package main

import (
	"bytes"
	"crypto/sha256"
	json "encoding/json/v2"
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

func TestScoopManifestIsDeterministic(t *testing.T) {
	rel := loadTestRelease(t)
	first, err := scoopManifest(rel)
	if err != nil {
		t.Fatalf("scoopManifest() error = %v", err)
	}
	second, err := scoopManifest(rel)
	if err != nil {
		t.Fatalf("scoopManifest() error = %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("scoopManifest() is not deterministic")
	}
	if !bytes.HasSuffix(first, []byte("\n")) {
		t.Fatal("scoopManifest() does not end with a newline")
	}
	var manifest map[string]any
	if err := json.Unmarshal(first, &manifest); err != nil {
		t.Fatalf("scoopManifest() wrote invalid JSON: %v", err)
	}
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

func TestRunRequiresFlags(t *testing.T) {
	if code := run(nil, &bytes.Buffer{}, &bytes.Buffer{}); code != 2 {
		t.Fatalf("run() = %d, want 2", code)
	}
	if code := run([]string{"-unknown"}, &bytes.Buffer{}, &bytes.Buffer{}); code != 2 {
		t.Fatalf("run() = %d, want 2 for an invalid flag", code)
	}
	if code := run([]string{"-h"}, &bytes.Buffer{}, &bytes.Buffer{}); code != 0 {
		t.Fatalf("run(-h) = %d, want 0", code)
	}
}

func TestRunBadChecksums(t *testing.T) {
	stderr := &bytes.Buffer{}
	code := run([]string{"-version", "0.2.0", "-checksums", filepath.Join(t.TempDir(), "missing")},
		&bytes.Buffer{}, stderr)
	if code != 1 {
		t.Fatalf("run() = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "update-packaging:") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRunWritesManifests(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"Formula", "bucket", filepath.Join("packaging", "aur")} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	checksums := writeFile(t, "checksums.txt", checksumsContent("0.2.0"))
	aur := writeFile(t, "go-fuck-0.2.0.tar.gz", "archive contents")

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := run([]string{
		"-version", "0.2.0",
		"-checksums", checksums,
		"-aur-source", aur,
		"-root", root,
	}, stdout, stderr)
	if code != 0 {
		t.Fatalf("run() = %d, stderr = %q", code, stderr.String())
	}

	for _, file := range []string{
		filepath.Join(root, "Formula", "go-fuck.rb"),
		filepath.Join(root, "bucket", "go-fuck.json"),
		filepath.Join(root, "packaging", "aur", "PKGBUILD"),
		filepath.Join(root, "packaging", "aur", ".SRCINFO"),
	} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		if !strings.Contains(string(data), "0.2.0") {
			t.Fatalf("%s does not contain the version: %q", file, string(data))
		}
	}
	if !strings.Contains(stdout.String(), "updated") {
		t.Fatalf("stdout = %q, want updated lines", stdout.String())
	}
}

func TestRunRefusesStaleAUR(t *testing.T) {
	root := t.TempDir()
	aurDir := filepath.Join(root, "packaging", "aur")
	if err := os.MkdirAll(aurDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(aurDir, "PKGBUILD"), []byte("pkgver=0.1.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	checksums := writeFile(t, "checksums.txt", checksumsContent("0.2.0"))
	stderr := &bytes.Buffer{}
	code := run([]string{"-version", "0.2.0", "-checksums", checksums, "-root", root},
		&bytes.Buffer{}, stderr)
	if code != 1 {
		t.Fatalf("run() = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "another version") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRunWriteFailure(t *testing.T) {
	root := writeFile(t, "root", "not a directory")
	checksums := writeFile(t, "checksums.txt", checksumsContent("0.2.0"))
	if code := run([]string{"-version", "0.2.0", "-checksums", checksums, "-root", root},
		&bytes.Buffer{}, &bytes.Buffer{}); code != 1 {
		t.Fatalf("run() = %d, want 1", code)
	}
}
