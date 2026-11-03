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
  version "0.10.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.10.0/go-fuck_0.10.0_darwin_arm64.tar.gz"
      sha256 "63a31523b48eef962b68354084a3b8b1324b0e149de7a54b4be94ecaec845e67"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.10.0/go-fuck_0.10.0_darwin_amd64.tar.gz"
      sha256 "34e31f973c0abe481cfcbdb81902f388d32ad8d098a5e0904edf2ca85bd3b60d"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.10.0/go-fuck_0.10.0_linux_arm64.tar.gz"
      sha256 "e33537ae9b3388f7632ac1417d8f189eec6affc63aab4a42fa4d2f81582395fe"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.10.0/go-fuck_0.10.0_linux_amd64.tar.gz"
      sha256 "358cd7c3aebaeb89ffe2affed46f2715a83ef4f5835645067ebf20b3f37d0add"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
