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
  version "2.29.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.29.0/go-fuck_2.29.0_darwin_arm64.tar.gz"
      sha256 "f1d579de401312f628a949138db32946b26c6c67acaaebc86800ffaed7693acd"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.29.0/go-fuck_2.29.0_darwin_amd64.tar.gz"
      sha256 "aa95caeca71ce18f5db6edab6881c1a9e2151149a522a07185de73b722940860"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.29.0/go-fuck_2.29.0_linux_arm64.tar.gz"
      sha256 "ced99a9167c5655fa0315adf48eb8b1680b19e75d118f3082fec3e3db7c9bca2"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.29.0/go-fuck_2.29.0_linux_amd64.tar.gz"
      sha256 "64c921879839fdabe5af40b6fc5156003be47f173ec20a747000d3111b1d015a"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
