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
  version "2.14.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.14.0/go-fuck_2.14.0_darwin_arm64.tar.gz"
      sha256 "bcd42bd6ac29a81d8b316843cf26ae833e4b43592ae0c77dadfe90a1be8a30fe"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.14.0/go-fuck_2.14.0_darwin_amd64.tar.gz"
      sha256 "6f8b11cea8493366416f60409194f929dd3c3c9ad05eb7af9d7b3c8ab906071f"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.14.0/go-fuck_2.14.0_linux_arm64.tar.gz"
      sha256 "563469c14e3a96fdc88dd9e64d736a5722ce59f43fcc28a3b7210ff1121fcc87"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.14.0/go-fuck_2.14.0_linux_amd64.tar.gz"
      sha256 "e060ec69485ccd56ea1877e9781a4a0e3383e6859de049979a6c0f552ced2e22"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
