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
  version "2.20.2"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.2/go-fuck_2.20.2_darwin_arm64.tar.gz"
      sha256 "b79a45d847415a74eafb6d0f7e095d073b86f66bea6c7761d53dce288b645adf"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.2/go-fuck_2.20.2_darwin_amd64.tar.gz"
      sha256 "b6bdfc520ac77418d795dfac3c1bceb5ae06380edee831d26375bb733c306535"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.2/go-fuck_2.20.2_linux_arm64.tar.gz"
      sha256 "b5be89bd26a05752761f2d505aa824ec2dace1a6413369b76f0dbbf33ee20b56"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.2/go-fuck_2.20.2_linux_amd64.tar.gz"
      sha256 "78c1d081725eb7786597c188910f87eb329efabcba420276688389fd2fdd87ca"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
