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
  version "2.31.15"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.15/go-fuck_2.31.15_darwin_arm64.tar.gz"
      sha256 "54bcd6b2d4baf8e2b8876cf9269753ecd10bfdd0bbf6e4b957041185bbd62c6d"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.15/go-fuck_2.31.15_darwin_amd64.tar.gz"
      sha256 "ec0306de1962babbf05524390a1c398071097215979daa87e8df9236efd066b5"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.15/go-fuck_2.31.15_linux_arm64.tar.gz"
      sha256 "463683eb95dc6f1816e421819814cadce707d9c21e3ee2d7657297f7736eae15"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.15/go-fuck_2.31.15_linux_amd64.tar.gz"
      sha256 "1ee8128e9af98a91a6e301c1de6581e12d55b2e163a98f941bd38c3cca6b8ac3"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
