# Homebrew formula for go-fuck.
#
# This repository doubles as a Homebrew tap, so users can install the tool with:
#
#   brew tap leafevans/go-fuck https://github.com/leafevans/go-fuck
#   brew install go-fuck
#
# The formula installs the prebuilt release archive for the current platform.
# Update it with "go run ./tools/update-packaging" when a new release is
# published (see packaging/README.md).
class GoFuck < Formula
  desc "Blazingly fast reimplementation of thefuck in Go"
  homepage "https://github.com/leafevans/go-fuck"
  version "2.31.9"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.9/go-fuck_2.31.9_darwin_arm64.tar.gz"
      sha256 "0914f28a17854ffd839ee065ccbe5393ae633f5006f5ca350305fa856c98ff2c"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.9/go-fuck_2.31.9_darwin_amd64.tar.gz"
      sha256 "73c6b8a533dc5fb3a1fad51829db77060896f106dfeb83da2a1cc20a3566389c"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.9/go-fuck_2.31.9_linux_arm64.tar.gz"
      sha256 "bc7fef11d218f0c4eccad226996796117a6f5ed0e9cce2f2f5b30f06de7a5606"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.9/go-fuck_2.31.9_linux_amd64.tar.gz"
      sha256 "bd50ecd93b3f8fde6c7abc96fa90f7886201659b12eaefeeb8fdd9c6848c21c8"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
