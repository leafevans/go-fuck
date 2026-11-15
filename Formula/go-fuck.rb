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
  version "2.31.2"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.2/go-fuck_2.31.2_darwin_arm64.tar.gz"
      sha256 "1387c91d95d426bed95202068fb0d76dc4d7973988ca41a54bdb2b1e3aed2156"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.2/go-fuck_2.31.2_darwin_amd64.tar.gz"
      sha256 "b1a9f7aedbc7c74667aee205fdf4be4562474d99a9ea6ff714752abafdc67d6f"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.2/go-fuck_2.31.2_linux_arm64.tar.gz"
      sha256 "877e1ed260223bc9596183d45fe04b451f625e8d750e0d6f15836680d5104039"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.2/go-fuck_2.31.2_linux_amd64.tar.gz"
      sha256 "5efd69b3f28ea869a6dfb9900005367c7f451da9d91100a8981f7e8b4dc36bdf"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
