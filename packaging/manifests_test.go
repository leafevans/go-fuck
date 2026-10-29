package packaging_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// manifestFiles lists the distribution manifests validated by the tests.
var manifestFiles = []string{
	"Formula/go-fuck.rb",
	"bucket/go-fuck.json",
	"packaging/winget/leafevans.gof.yaml",
	"packaging/winget/leafevans.gof.installer.yaml",
	"packaging/winget/leafevans.gof.locale.en-US.yaml",
}

// expectedAssets maps the six release assets to their archive extension.
var expectedAssets = map[string]string{
	"darwin_amd64":  "tar.gz",
	"darwin_arm64":  "tar.gz",
	"linux_amd64":   "tar.gz",
	"linux_arm64":   "tar.gz",
	"windows_amd64": "zip",
	"windows_arm64": "zip",
}

// splitLines splits content into lines without relying on escape sequences.
func splitLines(content string) []string {
	return strings.Split(content, string(rune(10)))
}

// unquote removes surrounding double quotes.
func unquote(value string) string {
	return strings.Trim(value, string(rune(34)))
}

// validHash reports whether hash is a lowercase SHA-256 hex string.
func validHash(hash string) bool {
	if len(hash) != 64 {
		return false
	}
	for _, character := range hash {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return false
		}
	}
	return true
}

// assetOf parses a release URL and returns the operating system and
// architecture key, for example linux_amd64.
func assetOf(rawURL, version, expectedExtension string) (string, error) {
	base := rawURL
	if index := strings.LastIndexByte(base, '/'); index >= 0 {
		base = base[index+1:]
	}
	parts := strings.Split(base, "_")
	if len(parts) != 4 || parts[0] != "go-fuck" {
		return "", fmt.Errorf("unexpected archive URL %q", rawURL)
	}
	if parts[1] != version {
		return "", fmt.Errorf("archive URL %q does not point at version %q", rawURL, version)
	}
	extension := ""
	switch {
	case strings.HasSuffix(parts[3], ".tar.gz"):
		extension = "tar.gz"
	case strings.HasSuffix(parts[3], ".zip"):
		extension = "zip"
	default:
		return "", fmt.Errorf("archive URL %q has an unknown extension", rawURL)
	}
	if extension != expectedExtension {
		return "", fmt.Errorf("archive URL %q has extension %q, want %q", rawURL, extension, expectedExtension)
	}
	architecture := strings.TrimSuffix(parts[3], "."+extension)
	return parts[2] + "_" + architecture, nil
}

// fieldValues returns all values of a line prefix in content.
func fieldValues(content, prefix string) []string {
	values := []string{}
	for _, line := range splitLines(content) {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			values = append(values, strings.TrimSpace(strings.TrimPrefix(line, prefix)))
		}
	}
	return values
}

// check validates the manifests below root.
func check(root string) error {
	problems := []string{}
	add := func(format string, args ...any) {
		problems = append(problems, fmt.Sprintf(format, args...))
	}

	formulaData, err := os.ReadFile(filepath.Join(root, "Formula", "go-fuck.rb"))
	if err != nil {
		return fmt.Errorf("read formula: %w", err)
	}
	formula := string(formulaData)

	version := ""
	for _, line := range splitLines(formula) {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "version ") {
			version = unquote(strings.TrimSpace(strings.TrimPrefix(line, "version ")))
			break
		}
	}
	if version == "" {
		return errors.New("formula does not declare a version")
	}

	urls := map[string]string{}
	hashes := map[string]bool{}
	for _, line := range splitLines(formula) {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "url "):
			rawURL := unquote(strings.TrimSpace(strings.TrimPrefix(line, "url ")))
			asset, err := assetOf(rawURL, version, "tar.gz")
			if err != nil {
				add("formula: %v", err)
				continue
			}
			urls[asset] = rawURL
		case strings.HasPrefix(line, "sha256 "):
			hash := unquote(strings.TrimSpace(strings.TrimPrefix(line, "sha256 ")))
			if !validHash(hash) {
				add("formula checksum %q is not a SHA-256 value", hash)
				continue
			}
			hashes[hash] = true
		}
	}
	if len(urls) != 4 {
		add("formula references %d archives, want 4", len(urls))
	}
	if len(hashes) != 4 {
		add("formula declares %d checksums, want 4", len(hashes))
	}

	scoopData, err := os.ReadFile(filepath.Join(root, "bucket", "go-fuck.json"))
	if err != nil {
		return fmt.Errorf("read scoop manifest: %w", err)
	}
	var scoop map[string]any
	if err := json.Unmarshal(scoopData, &scoop); err != nil {
		return fmt.Errorf("parse scoop manifest: %w", err)
	}
	if got, _ := scoop["version"].(string); got != version {
		add("scoop version is %q, formula version is %q", got, version)
	}
	architectures, _ := scoop["architecture"].(map[string]any)
	for key, asset := range map[string]string{"64bit": "windows_amd64", "arm64": "windows_arm64"} {
		entry, _ := architectures[key].(map[string]any)
		rawURL, _ := entry["url"].(string)
		hash, _ := entry["hash"].(string)
		parsed, err := assetOf(rawURL, version, "zip")
		if err != nil {
			add("scoop: %v", err)
		} else if parsed != asset {
			add("scoop references %q, want %q", parsed, asset)
		}
		if !validHash(hash) {
			add("scoop checksum %q is not a SHA-256 value", hash)
		}
		urls[asset] = rawURL
		hashes[hash] = true
	}

	wingetDir := filepath.Join(root, "packaging", "winget")
	wingetFiles := map[string]string{
		"version":   "leafevans.gof.yaml",
		"installer": "leafevans.gof.installer.yaml",
		"locale":    "leafevans.gof.locale.en-US.yaml",
	}
	winget := map[string]string{}
	for name, file := range wingetFiles {
		data, err := os.ReadFile(filepath.Join(wingetDir, file))
		if err != nil {
			return fmt.Errorf("read winget %s manifest: %w", name, err)
		}
		winget[name] = string(data)
		if versions := fieldValues(winget[name], "PackageVersion: "); len(versions) != 1 || versions[0] != version {
			add("winget %s manifest version is %q, formula version is %q", name, versions, version)
		}
	}

	installerURLs := fieldValues(winget["installer"], "InstallerUrl: ")
	installerHashes := fieldValues(winget["installer"], "InstallerSha256: ")
	if len(installerURLs) != 2 {
		add("winget installer references %d archives, want 2", len(installerURLs))
	}
	if len(installerHashes) != 2 {
		add("winget installer declares %d checksums, want 2", len(installerHashes))
	}
	windowsAssets := []string{"windows_amd64", "windows_arm64"}
	for index, rawURL := range installerURLs {
		asset := ""
		if index < len(windowsAssets) {
			asset = windowsAssets[index]
		}
		parsed, err := assetOf(rawURL, version, "zip")
		if err != nil {
			add("winget: %v", err)
		} else if parsed != asset {
			add("winget references %q, want %q", parsed, asset)
		}
		if index < len(installerHashes) {
			hash := installerHashes[index]
			if !validHash(hash) {
				add("winget checksum %q is not a SHA-256 value", hash)
			}
			hashes[hash] = true
		}
		urls[asset] = rawURL
	}

	if notes := fieldValues(winget["locale"], "ReleaseNotesUrl: "); len(notes) != 1 || !strings.HasSuffix(notes[0], "/v"+version) {
		add("winget release notes URL %q does not point at v%s", notes, version)
	}

	// winget-pkgs requires a professional command name, so the package is
	// published as leafevans.gof and installs the gof command.
	for name, content := range winget {
		if ids := fieldValues(content, "PackageIdentifier: "); len(ids) != 1 || ids[0] != "leafevans.gof" {
			add("winget %s PackageIdentifier = %q, want leafevans.gof", name, ids)
		}
	}
	if aliases := fieldValues(winget["installer"], "PortableCommandAlias: "); len(aliases) != 1 || aliases[0] != "gof" {
		add("winget PortableCommandAlias = %q, want gof", aliases)
	}
	if monikers := fieldValues(winget["locale"], "Moniker: "); len(monikers) != 1 || monikers[0] != "gof" {
		add("winget Moniker = %q, want gof", monikers)
	}

	missingAssets := []string{}
	for asset := range expectedAssets {
		if _, ok := urls[asset]; !ok {
			missingAssets = append(missingAssets, asset)
		}
	}
	if len(missingAssets) > 0 {
		add("manifests do not reference the assets: %s", strings.Join(missingAssets, ", "))
	}
	if len(hashes) != 6 {
		add("manifests carry %d distinct checksums, want 6", len(hashes))
	}

	if len(problems) > 0 {
		return fmt.Errorf("distribution manifests are inconsistent: %s", strings.Join(problems, "; "))
	}
	return nil
}

// copyManifests copies the committed manifests into a temporary repository.
func copyManifests(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, file := range manifestFiles {
		data, err := os.ReadFile(filepath.Join("..", file))
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		target := filepath.Join(root, file)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatalf("create directory for %s: %v", file, err)
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			t.Fatalf("write %s: %v", file, err)
		}
	}
	return root
}

func TestManifestsAreConsistent(t *testing.T) {
	if err := check(".."); err != nil {
		t.Fatalf("committed manifests are inconsistent: %v", err)
	}
}

func TestCheckDetectsVersionMismatch(t *testing.T) {
	root := copyManifests(t)
	path := filepath.Join(root, "bucket", "go-fuck.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	manifest["version"] = "9.9.9"
	updated, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, updated, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := check(root); err == nil {
		t.Fatal("check() should reject a version mismatch")
	}
}

func TestCheckDetectsMissingAsset(t *testing.T) {
	root := copyManifests(t)
	data, err := os.ReadFile(filepath.Join(root, "Formula", "go-fuck.rb"))
	if err != nil {
		t.Fatal(err)
	}
	lines := splitLines(string(data))
	for index, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "url ") {
			lines = append(lines[:index], lines[index+1:]...)
			break
		}
	}
	if err := os.WriteFile(filepath.Join(root, "Formula", "go-fuck.rb"), []byte(strings.Join(lines, string(rune(10)))), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := check(root); err == nil {
		t.Fatal("check() should reject a missing archive")
	}
}

func TestCheckDetectsInvalidChecksum(t *testing.T) {
	root := copyManifests(t)
	data, err := os.ReadFile(filepath.Join(root, "Formula", "go-fuck.rb"))
	if err != nil {
		t.Fatal(err)
	}
	lines := splitLines(string(data))
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "sha256 ") {
			lines[index] = strings.Replace(line, unquote(strings.TrimSpace(strings.TrimPrefix(trimmed, "sha256 "))), "not-a-checksum", 1)
			break
		}
	}
	if err := os.WriteFile(filepath.Join(root, "Formula", "go-fuck.rb"), []byte(strings.Join(lines, string(rune(10)))), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := check(root); err == nil {
		t.Fatal("check() should reject an invalid checksum")
	}
}
