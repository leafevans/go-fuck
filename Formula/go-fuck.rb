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
  version "2.31.3"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.3/go-fuck_2.31.3_darwin_arm64.tar.gz"
      sha256 "b2967a089a5224be3c4812ce12a4ed018bc325644c9ad66daa75753ecc50d3cd"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.3/go-fuck_2.31.3_darwin_amd64.tar.gz"
      sha256 "1b3bf44117da0c3116939571f0d3ae2a8cf553ea84e3f0d7d92b772ce1f389e0"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.3/go-fuck_2.31.3_linux_arm64.tar.gz"
      sha256 "9d07c42f4308b33a54fb37327fbff73262a1de6630c4f9edc6f5b7b1c9864287"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.3/go-fuck_2.31.3_linux_amd64.tar.gz"
      sha256 "221970dfbeecbc3197cbe423571072f462915c46b2c48d3286c27ed0d0952dcd"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
