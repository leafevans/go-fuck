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
  version "4.3.5"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.3.5/go-fuck_4.3.5_darwin_arm64.tar.gz"
      sha256 "3500347cfda379cba762f04a1b2092dd95801300bd55ed3915eaca3fe308a105"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.3.5/go-fuck_4.3.5_darwin_amd64.tar.gz"
      sha256 "11a48f01be1397f3c9ef28e9261d0c6f5801490c815012cc14541fc816c9d37f"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.3.5/go-fuck_4.3.5_linux_arm64.tar.gz"
      sha256 "dd07ddf3ff5ce3aa72f31a592dadd9b9c6029fc4c070c94c9d38faf82a69d8e0"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.3.5/go-fuck_4.3.5_linux_amd64.tar.gz"
      sha256 "d33c7387ddde5e8926b72d164a0f96dcc1b4c2837250d4bce73187de4a978060"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
