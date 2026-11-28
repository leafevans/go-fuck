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
  version "4.3.3"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.3.3/go-fuck_4.3.3_darwin_arm64.tar.gz"
      sha256 "988dade2f36eb1d405153e3917132c34bb36288808842417cc6068ac14b92283"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.3.3/go-fuck_4.3.3_darwin_amd64.tar.gz"
      sha256 "562c9e5404a74152a65f00fde348ce27ad541345028696d77da5c355d5a608d5"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.3.3/go-fuck_4.3.3_linux_arm64.tar.gz"
      sha256 "00eb0219df745032c5cab9c1836b75f1da01ce649c068ae71bb1b44a1b39eafe"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.3.3/go-fuck_4.3.3_linux_amd64.tar.gz"
      sha256 "1be04f886856b9faff7a22fef5172354e8bd4d0222e85aa45de65ed725578235"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
