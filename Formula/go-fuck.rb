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
  version "1.5.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.5.0/go-fuck_1.5.0_darwin_arm64.tar.gz"
      sha256 "23507ae5d9de15d5fbf4e999665c2fea6f680011ed4893f1a4daed0d5ab18d62"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.5.0/go-fuck_1.5.0_darwin_amd64.tar.gz"
      sha256 "d74eee22c1d39c78c43d836e5d1b8843c3ed8d9306dd7dfc9e4b611242c6af8d"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.5.0/go-fuck_1.5.0_linux_arm64.tar.gz"
      sha256 "a4b8aef43cdb3077401979b10bef37453cec6b2d83886d8ce7ba6e1cd9751916"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.5.0/go-fuck_1.5.0_linux_amd64.tar.gz"
      sha256 "b9372c35d13b00fcfe921d234ce4eece62752902b3287c267ee8f098320b059c"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
