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
  version "2.31.23"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.23/go-fuck_2.31.23_darwin_arm64.tar.gz"
      sha256 "111e7b30ec4f2345a1a6805b29d80a45888acba6830c813c1a5f492ee7bf7157"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.23/go-fuck_2.31.23_darwin_amd64.tar.gz"
      sha256 "81acd1cbac8226b3761b4c565dfef907447f5e0ed1ed313fc9c5a852c95a430b"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.23/go-fuck_2.31.23_linux_arm64.tar.gz"
      sha256 "0e9311b251c1e75abb6d97a40dad55afa22248511dc613a34554243361cfa3c0"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.23/go-fuck_2.31.23_linux_amd64.tar.gz"
      sha256 "ed30ae8fa41444d8b40e931fef73d55f4196f33ca33fe03178e00aa98c9dd1b2"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
