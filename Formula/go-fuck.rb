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
  version "2.5.2"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.5.2/go-fuck_2.5.2_darwin_arm64.tar.gz"
      sha256 "186dd67dce850c25b6b014d030082568f40f6e31b2b90e32b0ceeabad09a304d"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.5.2/go-fuck_2.5.2_darwin_amd64.tar.gz"
      sha256 "e98df6cc40504fc855dbef33d2e19982568ea0dafef551bf04e0111da9469165"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.5.2/go-fuck_2.5.2_linux_arm64.tar.gz"
      sha256 "6f5ef14b4120592ba2afa413d624b618fc6174ab665ba2b76d48330ae6232311"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.5.2/go-fuck_2.5.2_linux_amd64.tar.gz"
      sha256 "bc96539b9455811de6f2c26863f87a240d1e890bce32c1a02498965ef2636343"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
