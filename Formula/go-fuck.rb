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
  version "2.31.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.1/go-fuck_2.31.1_darwin_arm64.tar.gz"
      sha256 "0dfcb78e9316b4ff5ee0bd5a1c1952d11dc1f7c1ff45f63b477d305c913080a4"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.1/go-fuck_2.31.1_darwin_amd64.tar.gz"
      sha256 "ceae1eb92d569a08733224d5667b4193956470332c23706754483b129c663260"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.1/go-fuck_2.31.1_linux_arm64.tar.gz"
      sha256 "6f5131c42a4ff0d1469b26eb5e2463b86295449f1a310a7eee3ee17fd78d7ad2"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.1/go-fuck_2.31.1_linux_amd64.tar.gz"
      sha256 "2ce61f6d7a8f89229525721c45f1f8e37bfedf2393c47ebb224f4e89ff247ddc"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
