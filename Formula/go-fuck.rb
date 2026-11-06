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
  version "1.3.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.3.0/go-fuck_1.3.0_darwin_arm64.tar.gz"
      sha256 "441120f73b67c872e9c13e1bb6e8de861e25c3c4b3a9271cf62197bf04bf43ab"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.3.0/go-fuck_1.3.0_darwin_amd64.tar.gz"
      sha256 "f75297b60635ed2deb9fd44f8f2e99ce4bce40f6770704de61de7693234364e8"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.3.0/go-fuck_1.3.0_linux_arm64.tar.gz"
      sha256 "dc76da626295b395c39734ca27c8f7f692ae0eb08aea7796c85c96242aab12d1"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.3.0/go-fuck_1.3.0_linux_amd64.tar.gz"
      sha256 "b20ba8f3cc578c750b602720b096b58676a35b3051d3d07fe9f6359b8e6dcc54"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
