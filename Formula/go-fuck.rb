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
  version "0.3.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.3.1/go-fuck_0.3.1_darwin_arm64.tar.gz"
      sha256 "4a2de90c0dcef918330a3b6d3980984845eacd453542812c0942d5f1713c1212"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.3.1/go-fuck_0.3.1_darwin_amd64.tar.gz"
      sha256 "8fe89c58595ca4fdd6effa3eedfb7db801094c3fefa364aeeee28ce8b92bd995"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.3.1/go-fuck_0.3.1_linux_arm64.tar.gz"
      sha256 "f0484befdd1ba195e9edcd6e62a15ab0405b884b21f6e4e41aeeabf7795c12b3"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.3.1/go-fuck_0.3.1_linux_amd64.tar.gz"
      sha256 "00568fa1803a2c4a72a8eb751cbe95717f52f805fd6bb595cb10929ac0f4f568"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
