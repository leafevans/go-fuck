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
  version "2.31.21"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.21/go-fuck_2.31.21_darwin_arm64.tar.gz"
      sha256 "8148e5d09f7dc139771a9a9212cd89e0f05152874579fd0a1f56a22337b107d0"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.21/go-fuck_2.31.21_darwin_amd64.tar.gz"
      sha256 "941ded89962ee4284664c7902ccb963deb81667c51e4e932c8bcef1511dbba4a"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.21/go-fuck_2.31.21_linux_arm64.tar.gz"
      sha256 "58e7da1ede6f6507057941f5131e98aa974431232eef27a523478cb2ec0daabb"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.21/go-fuck_2.31.21_linux_amd64.tar.gz"
      sha256 "92407c3632ae738552dd40f917dc013d49820ac5c1e93a9d4cc442ba909f8897"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
