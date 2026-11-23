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
  version "3.0.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v3.0.1/go-fuck_3.0.1_darwin_arm64.tar.gz"
      sha256 "690a71fcfb87c489a91f529cc059f3e9132f3d6476d8449be247ce1fc591c415"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v3.0.1/go-fuck_3.0.1_darwin_amd64.tar.gz"
      sha256 "64a9ee5a629c324f764d72bffba18445f96ed321ba4c6d8408cabb700b65e601"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v3.0.1/go-fuck_3.0.1_linux_arm64.tar.gz"
      sha256 "edafa828c728c0858ad214319cdee18f158b25b325b80ec9ea762e0787b3b087"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v3.0.1/go-fuck_3.0.1_linux_amd64.tar.gz"
      sha256 "878d35edc1c160aff7c41a012c79f544570ec4f8d8e0d9902b10609a8f220fdc"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
