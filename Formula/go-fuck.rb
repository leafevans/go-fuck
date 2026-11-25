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
  version "4.2.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.2.0/go-fuck_4.2.0_darwin_arm64.tar.gz"
      sha256 "8f3bb646c4882936ca597383a1f539c9475740a1f1e947ed18b967fc0eb37bae"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.2.0/go-fuck_4.2.0_darwin_amd64.tar.gz"
      sha256 "9537396656de522d83e32054a757d9435d10170c51281a9ff1fb552b445efda3"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.2.0/go-fuck_4.2.0_linux_arm64.tar.gz"
      sha256 "3bdd375c0462eae83cc8da1a4f0785ff4b5011e0170b2d2495e7bc53ec975d21"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.2.0/go-fuck_4.2.0_linux_amd64.tar.gz"
      sha256 "1c2057c11d9174a2d9bdc88d79b16049f770c6c78a9f83a7a523fc0f68116668"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
