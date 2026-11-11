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
  version "2.0.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.0.0/go-fuck_2.0.0_darwin_arm64.tar.gz"
      sha256 "a8034053c40377161615f4091cda8d10f97f765c8f507e509b055e0241ccc96d"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.0.0/go-fuck_2.0.0_darwin_amd64.tar.gz"
      sha256 "4e267c441c3bf0365aaa7639b76fb0d3e03012f74c817fbff5439e85997bb363"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.0.0/go-fuck_2.0.0_linux_arm64.tar.gz"
      sha256 "ada68efac1b41c7a18193bc6ea333af06015113c8ab321c6abca31b2edb46aa6"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.0.0/go-fuck_2.0.0_linux_amd64.tar.gz"
      sha256 "7fd960609ce399bba85d345bb66f96f0389f100c3ab939ec4085b83fb2baf76d"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
