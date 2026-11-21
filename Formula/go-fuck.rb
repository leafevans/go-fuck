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
  version "2.34.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.34.0/go-fuck_2.34.0_darwin_arm64.tar.gz"
      sha256 "9b1d7d6ddae84e7bf41e29652b41a8ceeb3d81fdda2a9672787d4bf5e41633f6"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.34.0/go-fuck_2.34.0_darwin_amd64.tar.gz"
      sha256 "bec6f404c7ab0bec9c157662662680fe0f22e1b4a16b44810d86977e4fc879d3"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.34.0/go-fuck_2.34.0_linux_arm64.tar.gz"
      sha256 "766dee8e57bca01a34426c363e12b7e33442d00e67e216df9171044f6a9d9611"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.34.0/go-fuck_2.34.0_linux_amd64.tar.gz"
      sha256 "ed4973e2cb046347904b9579ecd0867ba6684007fafcf46b3e7009bbd59991b0"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
