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
  version "2.31.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.0/go-fuck_2.31.0_darwin_arm64.tar.gz"
      sha256 "27a3f261912aed0926d08ddcf76f771c1de343f641f2703d6c5a3c5b7455108f"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.0/go-fuck_2.31.0_darwin_amd64.tar.gz"
      sha256 "5cdc77559f532f60e70f0906258fe50c734fe6e3011f0e166d56b81bbdbaa885"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.0/go-fuck_2.31.0_linux_arm64.tar.gz"
      sha256 "4f5dde4ff279059ee9708979711a3747a91f536c5429078a925595eda6a62e70"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.0/go-fuck_2.31.0_linux_amd64.tar.gz"
      sha256 "573c0f0bb32100a183b950da3b64f59073095d500a6c792057d6efefb0766229"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
