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
  version "1.2.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.2.0/go-fuck_1.2.0_darwin_arm64.tar.gz"
      sha256 "79d1d150e96ef09d8b1769a97f7d1cdec0964bfb454780e7fcb7d6727cbaa9d8"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.2.0/go-fuck_1.2.0_darwin_amd64.tar.gz"
      sha256 "b69bc38d27dc28dceda8f0919d4a053e99c1e72c5f997efee2f637b3b7a317cf"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.2.0/go-fuck_1.2.0_linux_arm64.tar.gz"
      sha256 "330679e0ca2c78589dfa706f0718a8cd47a67c42d74cca8332e036f9b0955b57"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.2.0/go-fuck_1.2.0_linux_amd64.tar.gz"
      sha256 "13eefbae97bac2ca7b0a2f95159a913f317e6ca9bbb05bae45c42885cad17ea4"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
