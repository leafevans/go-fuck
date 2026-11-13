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
  version "2.22.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.22.0/go-fuck_2.22.0_darwin_arm64.tar.gz"
      sha256 "17868fd591b70534050d66b227d08e02885f56ec14f555f2c9e5154242fe75bc"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.22.0/go-fuck_2.22.0_darwin_amd64.tar.gz"
      sha256 "bb301bf7d05c90a58c1ef6d85d2fce068ffa11608c46d500ea03d81c593092d0"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.22.0/go-fuck_2.22.0_linux_arm64.tar.gz"
      sha256 "c674e593a855d4668ab66715ba3affdba936c5aaf492da99f35c075e20bb9dff"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.22.0/go-fuck_2.22.0_linux_amd64.tar.gz"
      sha256 "7780d1771aeaee5043ee961f9e14e4305b56ddad051aaec5d87b5fa7e0505714"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
