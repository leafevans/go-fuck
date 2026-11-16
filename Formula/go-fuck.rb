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
  version "2.31.5"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.5/go-fuck_2.31.5_darwin_arm64.tar.gz"
      sha256 "8d7338e1496ff2c84457415ad41b191fd308ae6916368663ff02618dbc1caf8a"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.5/go-fuck_2.31.5_darwin_amd64.tar.gz"
      sha256 "ec78e0b0b0173053163e36937ca9c9357e74bf4bbbf57d3e6e8856b10f3d8544"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.5/go-fuck_2.31.5_linux_arm64.tar.gz"
      sha256 "05a3d5132ddc2ec673c0d3691dcf9d0994a866371393138aec1847b1b5b70635"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.5/go-fuck_2.31.5_linux_amd64.tar.gz"
      sha256 "3bf8476de09d7b3d6130380bc4780842abb5a817645721cd5d9d76830bec578d"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
