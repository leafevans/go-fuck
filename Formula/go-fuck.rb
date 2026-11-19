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
  version "2.31.10"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.10/go-fuck_2.31.10_darwin_arm64.tar.gz"
      sha256 "d973afd5d2a973a76d761bfbf963a1b54302e750940ac70c234b9a927a7e447a"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.10/go-fuck_2.31.10_darwin_amd64.tar.gz"
      sha256 "f8d62572302cf75219c51fe1be44ddb7a698d66f2c81fd1deeec0155c0d3e8d6"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.10/go-fuck_2.31.10_linux_arm64.tar.gz"
      sha256 "52437b67af72d39f20168789e179b518c9301c7186acd3b4d718ee3ee351ba76"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.10/go-fuck_2.31.10_linux_amd64.tar.gz"
      sha256 "462c043cab65e61dfde6e10f91910d9f492ece156d7b35a0e0a1b7210eb95773"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
