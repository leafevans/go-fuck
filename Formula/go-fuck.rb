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
  version "0.8.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.8.0/go-fuck_0.8.0_darwin_arm64.tar.gz"
      sha256 "f372d3624146b53bc1bf7f84e88e0a6edb9754f82dd7255ac3a4365a354c4fb0"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.8.0/go-fuck_0.8.0_darwin_amd64.tar.gz"
      sha256 "5c702985ec707e9d28015c1712219b0da5e080f971eb8d81b9d575be9698c3aa"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.8.0/go-fuck_0.8.0_linux_arm64.tar.gz"
      sha256 "3c4a0a26980581596dc2409ffc252a33dc84e58dc2a4c051f982a16e83d9bf31"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.8.0/go-fuck_0.8.0_linux_amd64.tar.gz"
      sha256 "c22c0cfc1271e520ffcfeaec79720b3aab7a313c2e3215fe574da42ec480dd50"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
