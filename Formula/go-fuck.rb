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
  version "2.26.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.26.0/go-fuck_2.26.0_darwin_arm64.tar.gz"
      sha256 "ae6afbdc7e6d00c88294a4ab1cf7d29e98d4d08ab4636b39ddc80f7d87a46cc4"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.26.0/go-fuck_2.26.0_darwin_amd64.tar.gz"
      sha256 "b0a2fd69cb981e8f42d8feee20c21e7aa573c98caa15add05e42219d01b3d213"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.26.0/go-fuck_2.26.0_linux_arm64.tar.gz"
      sha256 "72df39cb2e35af017cb501bd56702c75d8987315a841bebc241b13dc8e0e284f"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.26.0/go-fuck_2.26.0_linux_amd64.tar.gz"
      sha256 "edf992baf9615423141ba581c6e0237e516f261ac0f6f78cb25546a713bfcf48"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
