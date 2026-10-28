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
  version "0.2.2"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.2.2/go-fuck_0.2.2_darwin_arm64.tar.gz"
      sha256 "440ec0e445f166516daba07c5459621e447e9fb7f19178956b4f54dc4f8990fa"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.2.2/go-fuck_0.2.2_darwin_amd64.tar.gz"
      sha256 "9669c81bb9aec01eb2c4876c05c366cf97daff47996a596ed58d3101a575999e"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.2.2/go-fuck_0.2.2_linux_arm64.tar.gz"
      sha256 "ce0b609aa3b1bc2a45078f0ff9cc17064aa4de5be2f9206436035723bd4e1c8e"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.2.2/go-fuck_0.2.2_linux_amd64.tar.gz"
      sha256 "52d1fe258e6419661df425ed2d848d56062ce75175f6392f4abafb1dbbddae80"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
