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
  version "2.15.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.15.0/go-fuck_2.15.0_darwin_arm64.tar.gz"
      sha256 "9815552ad7716234a72ef142ec2af7930c119503c62476d4280d6d27c1793b00"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.15.0/go-fuck_2.15.0_darwin_amd64.tar.gz"
      sha256 "2a129d9ceb6677c17ffd878a318406f411c3c555512ad57334f54ac13a4b2b6f"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.15.0/go-fuck_2.15.0_linux_arm64.tar.gz"
      sha256 "5182ba7728ed9606c590806f3fdcf00e069969ae78b0d6a259a088bfe54da6d7"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.15.0/go-fuck_2.15.0_linux_amd64.tar.gz"
      sha256 "913efc2ceebfb1ea52fc73d64d1bd596daabeea9cd6beb8ab5cb76058ce39b78"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
