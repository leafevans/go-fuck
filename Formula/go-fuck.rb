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
  version "0.6.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.6.0/go-fuck_0.6.0_darwin_arm64.tar.gz"
      sha256 "945c5f9b6dd81bdf7486dd3b4e0ca4cd496f3b7a290e6c35f15693a83bd12ddd"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.6.0/go-fuck_0.6.0_darwin_amd64.tar.gz"
      sha256 "86a0b6d232866c9637a0379529f59a074b8dd6936193ff7b8f9b0bdedab3cd49"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.6.0/go-fuck_0.6.0_linux_arm64.tar.gz"
      sha256 "e34b4c61bc1e5bf54d39c6b0973cd3c8ef392b573ccb06ac5a90d4865b8dbc4e"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.6.0/go-fuck_0.6.0_linux_amd64.tar.gz"
      sha256 "4d187582e4e7887c87d0e5717e555e73ffb048ba10dac6bb410d083c6d59f71a"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
