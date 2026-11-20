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
  version "2.31.16"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.16/go-fuck_2.31.16_darwin_arm64.tar.gz"
      sha256 "99d1bce04c836c942c313bd095859d08b549199747e09a06e20dcf23fc3d253a"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.16/go-fuck_2.31.16_darwin_amd64.tar.gz"
      sha256 "22a6523b9cc86a893bae633123c4cdc8afa65c74fa8676ad2c81c4b1da9c105b"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.16/go-fuck_2.31.16_linux_arm64.tar.gz"
      sha256 "9ab31601390134be9bc9558944989588399f253007593c9af3740028f05ac7b8"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.16/go-fuck_2.31.16_linux_amd64.tar.gz"
      sha256 "bf12c052ddec0d5d87c0114137584a3f22c7cc9a49ebcf910e7967be18cceef4"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
