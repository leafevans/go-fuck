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
  version "3.0.2"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v3.0.2/go-fuck_3.0.2_darwin_arm64.tar.gz"
      sha256 "d450ee2958f39101e3b02a47cb55e6e0431bc0639c9f5e9c1542ff4bcd41ff99"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v3.0.2/go-fuck_3.0.2_darwin_amd64.tar.gz"
      sha256 "a1afd36ededa62a1eca25b242129d990d62d107ba3501e04abcd07d6f314e615"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v3.0.2/go-fuck_3.0.2_linux_arm64.tar.gz"
      sha256 "870d89c2344fab12cb3f322b123114350e4b50e5fe3540570aebf398c2172b97"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v3.0.2/go-fuck_3.0.2_linux_amd64.tar.gz"
      sha256 "c9a45f2ad1d5bab8f212785236c3a1c42b35e5ea0c4e2dfb3b2eb7b507e6f0bc"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
