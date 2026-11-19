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
  version "2.31.11"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.11/go-fuck_2.31.11_darwin_arm64.tar.gz"
      sha256 "25dd2e10765d679bbb490675a61014fd31f503ea4647b30ed7b86807014c8deb"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.11/go-fuck_2.31.11_darwin_amd64.tar.gz"
      sha256 "00040502d3062e163e12afa28bbdb890516e864d40c26d27b5fa78cc4e6a56b1"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.11/go-fuck_2.31.11_linux_arm64.tar.gz"
      sha256 "783b844698c7ed63a70967508dde129fab2d5aa694b71ff51a6befff4ea0a4e5"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.11/go-fuck_2.31.11_linux_amd64.tar.gz"
      sha256 "f00e0b09ded5ae89124f5996145edf9bf8416a2eb8e50870f86b03aea36484fe"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
