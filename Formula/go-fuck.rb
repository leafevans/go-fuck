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
  version "2.31.19"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.19/go-fuck_2.31.19_darwin_arm64.tar.gz"
      sha256 "c94cb72e0db6e792cd35493f572d409463855f967fac3b57a3460e3b460a5a99"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.19/go-fuck_2.31.19_darwin_amd64.tar.gz"
      sha256 "a3eb409be89cc7c2e471e9273458b48a9aca0d297ee27a8f09d5d076265714b4"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.19/go-fuck_2.31.19_linux_arm64.tar.gz"
      sha256 "3b1c597d569677852b5b3d33e07c3fba3bb8fdace38da1015692d4855fb85266"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.19/go-fuck_2.31.19_linux_amd64.tar.gz"
      sha256 "697f77fac61ea4dd8e8e17cfcb42407ab830fbdf73379074cc525be05f69405c"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
