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
  version "4.3.2"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.3.2/go-fuck_4.3.2_darwin_arm64.tar.gz"
      sha256 "a3a4362ca1715b7f86fb10499ec5bc966fc4aade0e4c236136117b67b70fa245"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.3.2/go-fuck_4.3.2_darwin_amd64.tar.gz"
      sha256 "b1bde67d01d867e9a96db83c1e85fd81f3454e994de281c1119e1eaae17164dd"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.3.2/go-fuck_4.3.2_linux_arm64.tar.gz"
      sha256 "e3dfed1d4997ac532cda6a9f0a652bba3fbd404151dc13710a600b72cd7c7ef7"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.3.2/go-fuck_4.3.2_linux_amd64.tar.gz"
      sha256 "68e270c90bb265134c6d5926ca1577b85e4279538a048b1ed0b891c6ca424283"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
