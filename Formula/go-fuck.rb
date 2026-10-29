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
  version "0.3.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.3.0/go-fuck_0.3.0_darwin_arm64.tar.gz"
      sha256 "78ebae5ec29522fedee461c4e445c002f9d0d4174cee544bb155c81d0ed4913c"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.3.0/go-fuck_0.3.0_darwin_amd64.tar.gz"
      sha256 "b98479187a1b5de4c7f232b826fdef4cc10a36591a4acddadcf41ae94ceca1d7"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.3.0/go-fuck_0.3.0_linux_arm64.tar.gz"
      sha256 "3a2c37ec3455aa726cd030d8718c7dc49c5b961fefaf7be9f1669c263c479eba"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.3.0/go-fuck_0.3.0_linux_amd64.tar.gz"
      sha256 "2521394edfc54db35423d995ce3813370d8da720a25ba419afb0729f40998029"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
