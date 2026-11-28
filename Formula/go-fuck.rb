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
  version "4.3.4"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.3.4/go-fuck_4.3.4_darwin_arm64.tar.gz"
      sha256 "7e101f2ce420a89400057531e372458ce59eef57b974af36a5ae970f71df229a"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.3.4/go-fuck_4.3.4_darwin_amd64.tar.gz"
      sha256 "e2d7ee1223dca24e059af8cdb3de40ab16849f93d43bb69bbd0df935332a39ee"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.3.4/go-fuck_4.3.4_linux_arm64.tar.gz"
      sha256 "e4027e3d8b2bd0c520b3ef7cca973671da94c1b5851d5132e51d9440f5b6725d"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.3.4/go-fuck_4.3.4_linux_amd64.tar.gz"
      sha256 "f59feb3b0f9c358617b5fffc74d3d76f7f08459d88243eff4aadb142a46645ab"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
