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
  version "0.12.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.12.0/go-fuck_0.12.0_darwin_arm64.tar.gz"
      sha256 "a56136ead452d6cf8131902a8b3c3b4779613b1c0f9884fd50625663e0bc8b74"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.12.0/go-fuck_0.12.0_darwin_amd64.tar.gz"
      sha256 "68fbb35ff3d16918f39d7d5e43471ebd8210f4dcca89bf7aa05c870ad44e2594"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.12.0/go-fuck_0.12.0_linux_arm64.tar.gz"
      sha256 "8a80404d77ca68d6dff1de5502c048afea989a9be77945cf51bbaf1a8b0a91fe"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.12.0/go-fuck_0.12.0_linux_amd64.tar.gz"
      sha256 "7c16b1eb2717fc299a4c110e714798524facebe75f091cf97204fffe81566eb2"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
