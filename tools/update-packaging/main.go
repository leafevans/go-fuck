// Command update-packaging rewrites the distribution manifests for a release.
//
// It reads the checksums.txt asset of a GitHub release and updates the
// Homebrew formula and the Scoop manifest so that they point at the released
// archives. When -aur-source points at the GitHub tag
// archive of the release, the AUR PKGBUILD and .SRCINFO are regenerated too.
// The version is written without the "v" prefix; release assets are matched by
// name and validated to be complete.
//
// Usage:
//
//	go run ./tools/update-packaging -version 0.2.0 -checksums dist/checksums.txt -aur-source go-fuck-0.2.0.tar.gz
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json/jsontext"
	json "encoding/json/v2"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	projectName = "go-fuck"
	releaseBase = "https://github.com/leafevans/go-fuck/releases/download"
	archiveBase = "https://github.com/leafevans/go-fuck/archive/refs/tags"
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
	aurSource := flag.String("aur-source", "", "path of the GitHub tag archive (go-fuck-X.Y.Z.tar.gz) used by the AUR package")
	root := flag.String("root", ".", "repository root")
	flag.Parse()

	if *version == "" || *checksums == "" {
		fmt.Fprintln(os.Stderr, "usage: update-packaging -version VERSION -checksums CHECKSUMS [-aur-source ARCHIVE]")
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
		filepath.Join(*root, "Formula", "go-fuck.rb"):  formula(rel),
		filepath.Join(*root, "bucket", "go-fuck.json"): string(scoop),
	}

	aurDir := filepath.Join(*root, "packaging", "aur")
	if *aurSource == "" {
		// Keep the AUR files untouched without a source archive, but do not
		// silently leave them behind a new release.
		pkgbuild, err := os.ReadFile(filepath.Join(aurDir, "PKGBUILD"))
		if err == nil && !strings.Contains(string(pkgbuild), "pkgver="+rel.version+"\n") {
			fmt.Fprintln(os.Stderr, "update-packaging: packaging/aur is still at another version; pass -aur-source to update it")
			os.Exit(1)
		}
	} else {
		checksum, err := fileChecksum(*aurSource)
		if err != nil {
			fmt.Fprintln(os.Stderr, "update-packaging:", err)
			os.Exit(1)
		}
		files[filepath.Join(aurDir, "PKGBUILD")] = aurPKGBUILD(rel, checksum)
		files[filepath.Join(aurDir, ".SRCINFO")] = aurSrcinfo(rel, checksum)
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
	data, err := json.Marshal(manifest, jsontext.WithIndent("  "), json.Deterministic(true))
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

// aurPKGBUILDTemplate is the PKGBUILD of the AUR package. The release
// version and the checksum of the GitHub tag archive are filled in.
const aurPKGBUILDTemplate = `# Maintainer: LeafEvans <leafevans@foxmail.com>
pkgname=go-fuck
pkgver=%s
pkgrel=1
pkgdesc='A blazingly fast reimplementation of thefuck in Go'
arch=('x86_64')
url='https://github.com/leafevans/go-fuck'
license=('MIT')
makedepends=('go')
source=("$pkgname-$pkgver.tar.gz::$url/archive/refs/tags/v$pkgver.tar.gz")
sha256sums=('%s')

build() {
  cd "$pkgname-$pkgver"
  export CGO_ENABLED=0
  go build -trimpath \
    -ldflags "-s -w -X github.com/leafevans/go-fuck/internal/consts.Version=$pkgver" \
    -o go-fuck ./cmd/go-fuck
}

package() {
  cd "$pkgname-$pkgver"
  install -Dm755 go-fuck "$pkgdir/usr/bin/go-fuck"
  install -Dm644 LICENSE "$pkgdir/usr/share/licenses/$pkgname/LICENSE"
  install -Dm644 README.md "$pkgdir/usr/share/doc/$pkgname/README.md"
}
`

// aurSrcinfoTemplate is the .SRCINFO of the AUR package. The version, the
// archive name, the source URL and the checksum are filled in.
const aurSrcinfoTemplate = `pkgbase = go-fuck
	pkgdesc = A blazingly fast reimplementation of thefuck in Go
	pkgver = %s
	pkgrel = 1
	url = https://github.com/leafevans/go-fuck
	arch = x86_64
	license = MIT
	makedepends = go
	source = go-fuck-%s.tar.gz::%s/v%s.tar.gz
	sha256sums = %s

pkgname = go-fuck
`

// aurPKGBUILD returns the AUR PKGBUILD for the release.
func aurPKGBUILD(r *release, checksum string) string {
	return fmt.Sprintf(aurPKGBUILDTemplate, r.version, checksum)
}

// aurSrcinfo returns the AUR .SRCINFO for the release.
func aurSrcinfo(r *release, checksum string) string {
	return fmt.Sprintf(aurSrcinfoTemplate, r.version, r.version, archiveBase, r.version, checksum)
}

// fileChecksum returns the SHA-256 checksum of a file.
func fileChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("read AUR source archive: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("hash AUR source archive: %w", err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
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
