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
  version "2.13.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.13.1/go-fuck_2.13.1_darwin_arm64.tar.gz"
      sha256 "a91e1db1e80bcbd59309c61575ca5f42751812f55be2e869b5135c4ca3fc4f3d"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.13.1/go-fuck_2.13.1_darwin_amd64.tar.gz"
      sha256 "968d6bd4a84ec8a5bb75709ef26d39498388c4cacc7252ca1b80a35d02533d22"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.13.1/go-fuck_2.13.1_linux_arm64.tar.gz"
      sha256 "5ba78454a8c3e1dd61e0620aefa9774087d255f4facc5f352f8f0111db95d262"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.13.1/go-fuck_2.13.1_linux_amd64.tar.gz"
      sha256 "ebc16f94ebb96cff9530652d7e6f04c563e81976512eaecae3161c960e9f5d0c"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
