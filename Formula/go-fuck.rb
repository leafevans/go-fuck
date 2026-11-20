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
  version "2.31.20"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.20/go-fuck_2.31.20_darwin_arm64.tar.gz"
      sha256 "fd1440154a6f5e228dcba49e0e4b8b35c210c102dec883b4b793cc23185d8f69"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.20/go-fuck_2.31.20_darwin_amd64.tar.gz"
      sha256 "77db4a666cf849a820ab2eed02bd2d80146cf4c36a5f7d44bf07b264f0c896b6"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.20/go-fuck_2.31.20_linux_arm64.tar.gz"
      sha256 "bee050f938965c5627d019ba44a69a204be4d2417b90eb18e98f2e04f5f7758c"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.20/go-fuck_2.31.20_linux_amd64.tar.gz"
      sha256 "0a7d267968063bf7dd97511977705b247c2d77a91cd9e2b42a0cbf9133566b73"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
