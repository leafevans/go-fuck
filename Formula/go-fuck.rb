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
  version "2.5.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.5.1/go-fuck_2.5.1_darwin_arm64.tar.gz"
      sha256 "5c10cd0082580287797e950da233ed5430501c8b37d895dd31a0ae0cee9669b1"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.5.1/go-fuck_2.5.1_darwin_amd64.tar.gz"
      sha256 "1ffb911bc10e11d97e2c91628bbd6f743cfa6e8754a7d78f224a16c419519e1d"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.5.1/go-fuck_2.5.1_linux_arm64.tar.gz"
      sha256 "113c2a42aea938013898467e5e85fa7f701f85a1cf0a9947b5325c4e1a454c86"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.5.1/go-fuck_2.5.1_linux_amd64.tar.gz"
      sha256 "e1915f56b4b12a0a5fefd10dd9c53c0d22440b3a5cbdbcfb887305a1ed557cf2"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
