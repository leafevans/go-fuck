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
  version "2.19.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.19.1/go-fuck_2.19.1_darwin_arm64.tar.gz"
      sha256 "5e73b1922589d305dfdfe3b5de2fb1bbf4248de1d97c57f35f2c72b2cf23663b"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.19.1/go-fuck_2.19.1_darwin_amd64.tar.gz"
      sha256 "61b5b00d22c3f845ff398cc0d68e633472803326ed57cf4e5c1576830057e28f"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.19.1/go-fuck_2.19.1_linux_arm64.tar.gz"
      sha256 "bc018e005d2361feb3d67939f63977e6d913f53e423e52f866158102250f7ce4"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.19.1/go-fuck_2.19.1_linux_amd64.tar.gz"
      sha256 "181e27241d83c96afef338841f31927f9c873e9435997fcc0a43364f0c4a4303"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
