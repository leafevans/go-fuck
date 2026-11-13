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
  version "2.26.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.26.1/go-fuck_2.26.1_darwin_arm64.tar.gz"
      sha256 "6b51f2bdf76a6481287085943d2660d36a68c99367aac4e5d97309dc1ab395f7"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.26.1/go-fuck_2.26.1_darwin_amd64.tar.gz"
      sha256 "dbe7e455859fb020656d5d05282cc2c754c01c65615d4153850ff05d3e27b127"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.26.1/go-fuck_2.26.1_linux_arm64.tar.gz"
      sha256 "ad65bcbccb17853eabcff24eaff6397d2b3c20f20372652c6cbd86a0e52a9a02"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.26.1/go-fuck_2.26.1_linux_amd64.tar.gz"
      sha256 "8f817bc31d9374b488e9a218553455343e5d048ac514b2e7252e690ab006d877"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
