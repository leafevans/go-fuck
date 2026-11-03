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
  version "0.9.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.9.0/go-fuck_0.9.0_darwin_arm64.tar.gz"
      sha256 "d47ce32bc8659a1feaf9f7824180c8e2606b0ddacae64732bcd7cee0db5a6311"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.9.0/go-fuck_0.9.0_darwin_amd64.tar.gz"
      sha256 "a60188fc1b277074e0bea07a357673c2f6a9e1f83f1503d43cc903369253491e"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.9.0/go-fuck_0.9.0_linux_arm64.tar.gz"
      sha256 "d849a0e2c74f4143ec0eef8bfe08edfbf35158a6e8255e81ac85b7775dc8d743"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.9.0/go-fuck_0.9.0_linux_amd64.tar.gz"
      sha256 "2ff84710c0dfaea4930c4e12889d518c2deaa738b15cb921226a36768bf2a372"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
