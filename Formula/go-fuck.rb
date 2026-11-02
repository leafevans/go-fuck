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
  version "0.4.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.4.0/go-fuck_0.4.0_darwin_arm64.tar.gz"
      sha256 "14a59fa2e5c1ed352bb014112405ba140411871e6d27d219a8d2a245f2f33651"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.4.0/go-fuck_0.4.0_darwin_amd64.tar.gz"
      sha256 "6e443f5e29fda5a2c0b7578a97fa9c1a6f84277a5b757cea5fd7c81a87fa7bd5"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.4.0/go-fuck_0.4.0_linux_arm64.tar.gz"
      sha256 "7f3de0735fa6dc142022e809035ac0c4ee449fd32d2959b27719b7890533359a"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.4.0/go-fuck_0.4.0_linux_amd64.tar.gz"
      sha256 "b4a7941302aa3cc4dfc1bc3b587cb0936c87940092802c9a6c719f7e180f0baa"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
