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
  version "1.4.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.4.0/go-fuck_1.4.0_darwin_arm64.tar.gz"
      sha256 "005e3d9b583fcf62a1f1327d2232d3d03e2a39f140b506f454a94cced7be7a98"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.4.0/go-fuck_1.4.0_darwin_amd64.tar.gz"
      sha256 "4536b2dc14b7745b4e47f177bffb850504000069010a32bbb044945e45cc6830"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.4.0/go-fuck_1.4.0_linux_arm64.tar.gz"
      sha256 "f21d4ac8ee6cd688cf70f14d7779a68538b3c92be49980467e5495fe269e5236"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.4.0/go-fuck_1.4.0_linux_amd64.tar.gz"
      sha256 "815239ec725dca28d72620b68e74ad97416ac95b5d9b39bfc82e657351c898d3"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
