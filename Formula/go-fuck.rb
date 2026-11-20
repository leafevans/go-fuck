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
  version "2.31.22"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.22/go-fuck_2.31.22_darwin_arm64.tar.gz"
      sha256 "537de9e30d743933bc2054ad94694653c2edf705dfa166fd1071bc407131d554"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.22/go-fuck_2.31.22_darwin_amd64.tar.gz"
      sha256 "694ab724085436337bfeb2859cec08805d497e925fec1def4a8a937f6a25adc4"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.22/go-fuck_2.31.22_linux_arm64.tar.gz"
      sha256 "6e38458549b4102659241352b1ce5040a473b652557dc80cde2a2840b44b6cd6"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.22/go-fuck_2.31.22_linux_amd64.tar.gz"
      sha256 "398db0f17dfe34074e07588e6a24773645142857eeca532f430be02225faa88c"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
