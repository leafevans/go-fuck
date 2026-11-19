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
  version "2.31.6"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.6/go-fuck_2.31.6_darwin_arm64.tar.gz"
      sha256 "5a294a51c97f5df28145a333de237388d81a82e6c18bbf1f8be64c9a28f2ca30"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.6/go-fuck_2.31.6_darwin_amd64.tar.gz"
      sha256 "2b2f24ce11fddf0e687fb21e3ef1349b12d2e6b9b6e057b0cc49e4986d724def"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.6/go-fuck_2.31.6_linux_arm64.tar.gz"
      sha256 "e481a60e34117992ffbfe03e67221c77d2df37b4ce44ab3644a71e3bf0848fc5"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.6/go-fuck_2.31.6_linux_amd64.tar.gz"
      sha256 "fbd258b664ec024e486b0a088320ed7b700ccf730e6ccace69df88003daee064"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
