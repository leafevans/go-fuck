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
  version "2.7.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.7.0/go-fuck_2.7.0_darwin_arm64.tar.gz"
      sha256 "f04174a88c21c369f608a93bac4594ef4c9323b0924ee77f0a5ecb6118fc5f5c"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.7.0/go-fuck_2.7.0_darwin_amd64.tar.gz"
      sha256 "d0cc0af96341afef0b6cc2fb36f5b4d509f06c36ee4ce27112ca62d0712d2333"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.7.0/go-fuck_2.7.0_linux_arm64.tar.gz"
      sha256 "0d518b8eea3e9efa2d00d4fce78b8241228a549c47cdd5a459ddbb29b68511e7"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.7.0/go-fuck_2.7.0_linux_amd64.tar.gz"
      sha256 "7c9c5e247225f83bd8ec54367e25022992be3ee9c8178d3c0c830c0ca36218ef"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
