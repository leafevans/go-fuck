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
  version "1.8.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.8.1/go-fuck_1.8.1_darwin_arm64.tar.gz"
      sha256 "657912281d3ee46452e9003f96f51e59dd22bdb59a01863f185b32454b924cd8"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.8.1/go-fuck_1.8.1_darwin_amd64.tar.gz"
      sha256 "052b4dfbd4ef5d0a07988ac0aa0882d0b68f0a5b4409dad40d318b9913479403"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.8.1/go-fuck_1.8.1_linux_arm64.tar.gz"
      sha256 "62911268da97d4260251e9c2688e734991432be71526724bcf8623ecaff2f197"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.8.1/go-fuck_1.8.1_linux_amd64.tar.gz"
      sha256 "93874dd0d64bb4e79cc8fe6a90cd529fd8f8ca06ead7c1011e471a234b890f15"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
