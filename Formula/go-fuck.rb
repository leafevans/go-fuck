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
  version "2.35.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.35.0/go-fuck_2.35.0_darwin_arm64.tar.gz"
      sha256 "7565815dcef773d1ee687daec29517a809043ba2ea1900629e8ffe2d5fc30cd6"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.35.0/go-fuck_2.35.0_darwin_amd64.tar.gz"
      sha256 "98856feff81075d081dcecdeab4b0db9ab7f3cf39a8d80234a5216a0ec119227"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.35.0/go-fuck_2.35.0_linux_arm64.tar.gz"
      sha256 "32fead1aa823b52e494539ff7267cebfac5ce22f46abd67845e867ca794828fb"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.35.0/go-fuck_2.35.0_linux_amd64.tar.gz"
      sha256 "32666bccc8742114f2f509d60f97b18d6df6fc4753a567e3c88cac9ec600b2b6"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
