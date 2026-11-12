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
  version "2.20.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.1/go-fuck_2.20.1_darwin_arm64.tar.gz"
      sha256 "f73db43be016905eced53196f7d4b1d082fcd09359d14220a3c9b43da90a0532"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.1/go-fuck_2.20.1_darwin_amd64.tar.gz"
      sha256 "3324dd34fde6c43137d73225518ecc5e3714ca7cbe77d514008d0c56b030ccbd"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.1/go-fuck_2.20.1_linux_arm64.tar.gz"
      sha256 "c52270a815cb538e4e864efdab56f5ce4a618330e8378308d9279d4786bc4cbd"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.1/go-fuck_2.20.1_linux_amd64.tar.gz"
      sha256 "5557ab587b306dcd9416428974f1874e89883b7b732900c0258cb9b51218a08a"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
