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
  version "0.2.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.2.1/go-fuck_0.2.1_darwin_arm64.tar.gz"
      sha256 "3f2a8a189a8725d381db4469c98ef1ca337e56ff8774e1afa903d1bd22167677"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.2.1/go-fuck_0.2.1_darwin_amd64.tar.gz"
      sha256 "24df425cea9846f548cc5f15596af4a77b7433a6890d6b1056f8418a5d594310"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.2.1/go-fuck_0.2.1_linux_arm64.tar.gz"
      sha256 "9a74adf5314455409da2d5174b0259e288ab37d912568d95085ea93343bc4069"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.2.1/go-fuck_0.2.1_linux_amd64.tar.gz"
      sha256 "b2f51936a0303a4509da2cc06ccf3ab1199e50638263bc93d7fd92ab172db350"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
