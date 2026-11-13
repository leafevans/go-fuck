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
  version "2.21.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.21.0/go-fuck_2.21.0_darwin_arm64.tar.gz"
      sha256 "a37747ea84e6261f835e01f3f3055f8824d80ede5f6b69354169eabb52bcecfc"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.21.0/go-fuck_2.21.0_darwin_amd64.tar.gz"
      sha256 "27f1a3d060d62b08e8c1b5db0a084097b08ae395abf2bc1b68b1aa020cec4a3f"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.21.0/go-fuck_2.21.0_linux_arm64.tar.gz"
      sha256 "c96603506abe10daeac31c385daff9c92ce056b959297a6f4b34c4b50215b6dc"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.21.0/go-fuck_2.21.0_linux_amd64.tar.gz"
      sha256 "f43d8f20d738dc80b45338490a895c1346277f7070b15feda1e8801d51eb5926"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
