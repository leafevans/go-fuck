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
  version "2.2.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.2.0/go-fuck_2.2.0_darwin_arm64.tar.gz"
      sha256 "c6e01eb73d3862d5083c8c6231f4617ea88f905ac978dc84e0c63998f09d4f92"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.2.0/go-fuck_2.2.0_darwin_amd64.tar.gz"
      sha256 "e9872b23d51e5a6ee0d565f5fe73e3fcd0eebe609a23f40e32fc04cb088c2de3"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.2.0/go-fuck_2.2.0_linux_arm64.tar.gz"
      sha256 "98cdcdb3c42da380708539c65040d6554f2c9e9be91904806a270bf08a0ea053"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.2.0/go-fuck_2.2.0_linux_amd64.tar.gz"
      sha256 "02fdfbfd1ed47c1e0720a37ae092e98ba5b55e62dab83016faa5aae4f74c3921"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
