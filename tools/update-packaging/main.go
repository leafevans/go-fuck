// Command update-packaging rewrites the distribution manifests for a release.
//
// It reads the checksums.txt asset of a GitHub release and updates the
// Homebrew formula, the Scoop manifest and the winget manifests so that they
// point at the released archives. The version is written without the "v"
// prefix; release assets are matched by name and validated to be complete.
//
// Usage:
//
//	go run ./tools/update-packaging -version 0.2.0 -checksums dist/checksums.txt
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	projectName = "go-fuck"
	releaseBase = "https://github.com/leafevans/go-fuck/releases/download"
)

// checksumLine matches one line of a GoReleaser checksums file.
var checksumLine = regexp.MustCompile("^([0-9a-f]{64})\\s+" + projectName + "_([0-9A-Za-z.]+)_([a-z0-9]+)_([a-z0-9]+)\\.(tar\\.gz|zip)$")

// expectedAssets lists the checksum keys that every release must provide.
var expectedAssets = []string{
	"darwin_amd64",
	"darwin_arm64",
	"linux_amd64",
	"linux_arm64",
	"windows_amd64",
	"windows_arm64",
}

// release describes the assets of one release.
type release struct {
	version string
	hashes  map[string]string
}

func main() {
	version := flag.String("version", "", "release version, for example 0.2.0 (a leading v is accepted)")
	checksums := flag.String("checksums", "", "path of the checksums.txt asset of the release")
	root := flag.String("root", ".", "repository root")
	flag.Parse()

	if *version == "" || *checksums == "" {
		fmt.Fprintln(os.Stderr, "usage: update-packaging -version VERSION -checksums CHECKSUMS")
		flag.PrintDefaults()
		os.Exit(2)
	}

	rel, err := loadRelease(*checksums, strings.TrimPrefix(*version, "v"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "update-packaging:", err)
		os.Exit(1)
	}

	scoop, err := scoopManifest(rel)
	if err != nil {
		fmt.Fprintln(os.Stderr, "update-packaging:", err)
		os.Exit(1)
	}

	files := map[string]string{
		filepath.Join(*root, "Formula", "go-fuck.rb"):                                      formula(rel),
		filepath.Join(*root, "bucket", "go-fuck.json"):                                     string(scoop),
		filepath.Join(*root, "packaging", "winget", "leafevans.go-fuck.yaml"):              wingetVersionManifest(rel),
		filepath.Join(*root, "packaging", "winget", "leafevans.go-fuck.installer.yaml"):    wingetInstallerManifest(rel),
		filepath.Join(*root, "packaging", "winget", "leafevans.go-fuck.locale.en-US.yaml"): wingetLocaleManifest(rel),
	}

	for _, path := range sortedKeys(files) {
		if err := os.WriteFile(path, []byte(files[path]), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "update-packaging:", err)
			os.Exit(1)
		}
		fmt.Println("updated", path)
	}
}

// loadRelease parses and validates the checksums of a release.
func loadRelease(path, version string) (*release, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read checksums: %w", err)
	}

	rel := &release{version: version, hashes: map[string]string{}}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		match := checksumLine.FindStringSubmatch(line)
		if match == nil {
			return nil, fmt.Errorf("unexpected checksums line %q", line)
		}
		if match[2] != version {
			return nil, fmt.Errorf("checksums are for version %s, want %s", match[2], version)
		}
		rel.hashes[match[3]+"_"+match[4]] = match[1]
	}

	for _, key := range expectedAssets {
		if rel.hashes[key] == "" {
			return nil, fmt.Errorf("checksums are missing the %s asset", key)
		}
	}
	return rel, nil
}

// asset returns the release asset filename of an operating system and
// architecture.
func (r *release) asset(operatingSystem, architecture string) string {
	extension := ".tar.gz"
	if operatingSystem == "windows" {
		extension = ".zip"
	}
	return fmt.Sprintf("%s_%s_%s_%s%s", projectName, r.version, operatingSystem, architecture, extension)
}

// url returns the download URL of an operating system and architecture.
func (r *release) url(operatingSystem, architecture string) string {
	return releaseBase + "/v" + r.version + "/" + r.asset(operatingSystem, architecture)
}

// formula returns the Homebrew formula for the release.
func formula(r *release) string {
	var b strings.Builder
	b.WriteString("# Homebrew formula for go-fuck.\n")
	b.WriteString("#\n")
	b.WriteString("# This repository doubles as a Homebrew tap, so users can install the tool with:\n")
	b.WriteString("#\n")
	b.WriteString("#   brew tap leafevans/go-fuck https://github.com/leafevans/go-fuck\n")
	b.WriteString("#   brew install go-fuck\n")
	b.WriteString("#\n")
	b.WriteString("# The formula installs the prebuilt release archive for the current platform.\n")
	b.WriteString("# Update it with \"go run ./tools/update-packaging\" when a new release is\n")
	b.WriteString("# published (see packaging/README.md).\n")
	b.WriteString("class GoFuck < Formula\n")
	b.WriteString("  desc \"Blazingly fast reimplementation of thefuck in Go\"\n")
	b.WriteString("  homepage \"https://github.com/leafevans/go-fuck\"\n")
	fmt.Fprintf(&b, "  version %q\n", r.version)
	b.WriteString("  license \"MIT\"\n\n")
	b.WriteString("  on_macos do\n")
	fmt.Fprintf(&b, "    if Hardware::CPU.arm?\n      url %q\n      sha256 %q\n    else\n      url %q\n      sha256 %q\n    end\n",
		r.url("darwin", "arm64"), r.hashes["darwin_arm64"],
		r.url("darwin", "amd64"), r.hashes["darwin_amd64"])
	b.WriteString("  end\n\n")
	b.WriteString("  on_linux do\n")
	fmt.Fprintf(&b, "    if Hardware::CPU.arm?\n      url %q\n      sha256 %q\n    else\n      url %q\n      sha256 %q\n    end\n",
		r.url("linux", "arm64"), r.hashes["linux_arm64"],
		r.url("linux", "amd64"), r.hashes["linux_amd64"])
	b.WriteString("  end\n\n")
	b.WriteString("  def install\n    bin.install \"go-fuck\"\n  end\n\n")
	b.WriteString("  test do\n    assert_match \"go-fuck #{version}\", shell_output(\"#{bin}/go-fuck --version\")\n  end\n")
	b.WriteString("end\n")
	return b.String()
}

// scoopManifest returns the Scoop bucket manifest for the release.
func scoopManifest(r *release) ([]byte, error) {
	manifest := map[string]any{
		"$schema":     "https://raw.githubusercontent.com/ScoopInstaller/Scoop/master/schema.json",
		"version":     r.version,
		"description": "Blazingly fast reimplementation of thefuck in Go.",
		"homepage":    "https://github.com/leafevans/go-fuck",
		"license":     "MIT",
		"architecture": map[string]any{
			"64bit": map[string]any{
				"url":  r.url("windows", "amd64"),
				"hash": r.hashes["windows_amd64"],
			},
			"arm64": map[string]any{
				"url":  r.url("windows", "arm64"),
				"hash": r.hashes["windows_arm64"],
			},
		},
		"bin": "go-fuck.exe",
		"checkver": map[string]any{
			"github": "https://github.com/leafevans/go-fuck",
		},
		"autoupdate": map[string]any{
			"architecture": map[string]any{
				"64bit": map[string]any{
					"url": releaseBase + "/v$version/" + projectName + "_$version_windows_amd64.zip",
				},
				"arm64": map[string]any{
					"url": releaseBase + "/v$version/" + projectName + "_$version_windows_arm64.zip",
				},
			},
		},
	}
	return json.MarshalIndent(manifest, "", "  ")
}

// wingetVersionManifest returns the winget version manifest for the release.
func wingetVersionManifest(r *release) string {
	return "# yaml-language-server: $schema=https://aka.ms/winget-manifest.version.1.12.0.schema.json\n" +
		"# winget version manifest for go-fuck. Generated by tools/update-packaging.\n" +
		"PackageIdentifier: leafevans.go-fuck\n" +
		"PackageVersion: " + r.version + "\n" +
		"DefaultLocale: en-US\n" +
		"ManifestType: version\n" +
		"ManifestVersion: 1.12.0\n"
}

// wingetInstallerManifest returns the winget installer manifest for the release.
func wingetInstallerManifest(r *release) string {
	return "# yaml-language-server: $schema=https://aka.ms/winget-manifest.installer.1.12.0.schema.json\n" +
		"# winget installer manifest for go-fuck. Generated by tools/update-packaging.\n" +
		"PackageIdentifier: leafevans.go-fuck\n" +
		"PackageVersion: " + r.version + "\n" +
		"InstallerType: zip\n" +
		"NestedInstallerType: portable\n" +
		"NestedInstallerFiles:\n" +
		"  - RelativeFilePath: go-fuck.exe\n" +
		"    PortableCommandAlias: go-fuck\n" +
		"Installers:\n" +
		"  - Architecture: x64\n" +
		"    InstallerUrl: " + r.url("windows", "amd64") + "\n" +
		"    InstallerSha256: " + r.hashes["windows_amd64"] + "\n" +
		"  - Architecture: arm64\n" +
		"    InstallerUrl: " + r.url("windows", "arm64") + "\n" +
		"    InstallerSha256: " + r.hashes["windows_arm64"] + "\n" +
		"ManifestType: installer\n" +
		"ManifestVersion: 1.12.0\n"
}

// wingetLocaleManifest returns the winget default locale manifest.
func wingetLocaleManifest(r *release) string {
	return "# yaml-language-server: $schema=https://aka.ms/winget-manifest.defaultLocale.1.12.0.schema.json\n" +
		"# winget default locale manifest for go-fuck. Generated by tools/update-packaging.\n" +
		"PackageIdentifier: leafevans.go-fuck\n" +
		"PackageVersion: " + r.version + "\n" +
		"PackageLocale: en-US\n" +
		"Publisher: LeafEvans\n" +
		"PublisherUrl: https://github.com/leafevans\n" +
		"PublisherSupportUrl: https://github.com/leafevans/go-fuck/issues\n" +
		"Author: LeafEvans\n" +
		"PackageName: go-fuck\n" +
		"PackageUrl: https://github.com/leafevans/go-fuck\n" +
		"License: MIT\n" +
		"LicenseUrl: https://github.com/leafevans/go-fuck/blob/main/LICENSE\n" +
		"Copyright: Copyright (c) 2026 橘星奏\n" +
		"ShortDescription: Blazingly fast reimplementation of thefuck in Go\n" +
		"Description: go-fuck fixes the command you just typed and got wrong. It re-runs the failed command to capture its output, matches one of the built-in rules and offers the corrected command for execution in your shell.\n" +
		"Moniker: go-fuck\n" +
		"Tags:\n" +
		"  - cli\n" +
		"  - shell\n" +
		"  - correction\n" +
		"  - productivity\n" +
		"  - thefuck\n" +
		"  - go\n" +
		"ReleaseNotesUrl: https://github.com/leafevans/go-fuck/releases/tag/v" + r.version + "\n" +
		"ManifestType: defaultLocale\n" +
		"ManifestVersion: 1.12.0\n"
}

// sortedKeys returns the keys of a string map in lexicographic order.
func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}
