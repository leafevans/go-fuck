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
  version "2.6.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.6.0/go-fuck_2.6.0_darwin_arm64.tar.gz"
      sha256 "50046f1b5553a1624487cefd403bb8a6ed9d0ff328c08d8d7573276c7a2945a5"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.6.0/go-fuck_2.6.0_darwin_amd64.tar.gz"
      sha256 "b15e96f8ae700c9a4f0bcf4a75fe0614cd54596416f2f98f0f8dbfe8d004a7ce"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.6.0/go-fuck_2.6.0_linux_arm64.tar.gz"
      sha256 "b2406c4d0ab3c4041002bd631f24bf10b4198ce48d976df57914b3012dd40846"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.6.0/go-fuck_2.6.0_linux_amd64.tar.gz"
      sha256 "142b1b1199a2267418ba6563537c825799c35a8d6da80b03f15f312db49b548e"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
