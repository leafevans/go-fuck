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
  version "2.20.4"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.4/go-fuck_2.20.4_darwin_arm64.tar.gz"
      sha256 "3d1f61b9710c1bb239814cce30cfe9febeaba4094ead19ecd46b5d4e4b55a9ca"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.4/go-fuck_2.20.4_darwin_amd64.tar.gz"
      sha256 "199654755d9d93068c8350daab6db1022e8ad4004d5f2b28ec93bae08b130141"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.4/go-fuck_2.20.4_linux_arm64.tar.gz"
      sha256 "58611daeb82a1141da2fd78edf1bc9d5f997243facf26f45909ae4d53614b7cb"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.4/go-fuck_2.20.4_linux_amd64.tar.gz"
      sha256 "85aa6ee4086fbceb47e6a651b743500d95a03a91790d842e878e2ee46325ae19"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
