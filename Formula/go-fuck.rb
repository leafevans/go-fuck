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
  version "2.35.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.35.1/go-fuck_2.35.1_darwin_arm64.tar.gz"
      sha256 "fc47ecb728ef8e7feaccc0092dce6fb8eee42981a9aa46b5f097cddb513d9975"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.35.1/go-fuck_2.35.1_darwin_amd64.tar.gz"
      sha256 "01d43d88ce4b37466e8360197e68cd9ed30e7ec2efa70015435b7ed00fafa112"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.35.1/go-fuck_2.35.1_linux_arm64.tar.gz"
      sha256 "c3abc0531d85e78fde5bb1e06b21ca91b927da5d2c8b0a6a9c5aa28f1dbcf22c"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.35.1/go-fuck_2.35.1_linux_amd64.tar.gz"
      sha256 "f821608be4bb423c02d01ff3a6d43fdc918ebbd4a1497e047433135bca1dafd3"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
