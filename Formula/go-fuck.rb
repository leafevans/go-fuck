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
  version "4.3.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.3.0/go-fuck_4.3.0_darwin_arm64.tar.gz"
      sha256 "fc0b715e3034fc4801aacbd62c849baffbb885cd038edd3936e700e3238f6537"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.3.0/go-fuck_4.3.0_darwin_amd64.tar.gz"
      sha256 "56f3ece532b618c151045fd8d5d1973a8313ce7f0cd1f92197d004b343d2b949"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.3.0/go-fuck_4.3.0_linux_arm64.tar.gz"
      sha256 "9b9c929f0b30d8701f6694ae73ebf3cd2dbfc732d072979e6732c5507d30395d"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.3.0/go-fuck_4.3.0_linux_amd64.tar.gz"
      sha256 "0eebf009c916f2ad6c9c42d83b078184feb29a5cf443b0fdbea51f877106e252"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
