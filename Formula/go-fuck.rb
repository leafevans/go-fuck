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
  version "1.9.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.9.0/go-fuck_1.9.0_darwin_arm64.tar.gz"
      sha256 "5e1582eaf45b214f593aaef9972b13229f2d359f5e1342d081d5ca96eeffea41"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.9.0/go-fuck_1.9.0_darwin_amd64.tar.gz"
      sha256 "5ec4d49e46a4bdf7dae5115d63f9f4f8dacc7f99f54cdc838f9d11f43035805e"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.9.0/go-fuck_1.9.0_linux_arm64.tar.gz"
      sha256 "ec8b807b4003de06bcae8660725b8f8a36c44b896b01e140f6736ba8b4dab89b"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.9.0/go-fuck_1.9.0_linux_amd64.tar.gz"
      sha256 "686ab23896275ac25eb8c58a07ec7c4ed03ce4dad50c59fec028999c56960438"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
