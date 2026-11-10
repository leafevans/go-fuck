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
  version "1.12.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.12.0/go-fuck_1.12.0_darwin_arm64.tar.gz"
      sha256 "f7a6aa320d01d957d543acd8052fbd151972c13cf351e5c5a05f0b0fd1c2b183"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.12.0/go-fuck_1.12.0_darwin_amd64.tar.gz"
      sha256 "d92c5746d1fd4df477c5bebf21d4e84017984546e1d2a72c157a0a96e1f470bc"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.12.0/go-fuck_1.12.0_linux_arm64.tar.gz"
      sha256 "1b467660cd1a52de531286751bd98136e1f045a209a73758ec552404a68f1b4c"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.12.0/go-fuck_1.12.0_linux_amd64.tar.gz"
      sha256 "46464c35513aaca51669799ebb751f595881e3dd97064261c8afd352eb7d68ca"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
