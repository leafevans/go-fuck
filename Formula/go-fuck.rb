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
  version "2.5.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.5.0/go-fuck_2.5.0_darwin_arm64.tar.gz"
      sha256 "fa3dfb52933706ad88711a586b59e8ebc2d0f63f7f99798b0e76c56066a49adc"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.5.0/go-fuck_2.5.0_darwin_amd64.tar.gz"
      sha256 "95a75aa90e68b24a9cdea0165d0bfeeba04009ae6ed22712370144a62f6037e3"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.5.0/go-fuck_2.5.0_linux_arm64.tar.gz"
      sha256 "49c04327399df812f23339b9ad9a09586fea077752c7828fae5d15c55e219140"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.5.0/go-fuck_2.5.0_linux_amd64.tar.gz"
      sha256 "97afc48a6ba4d7fdea961e5d2ff1ccec586e1c5c6b7bb98a519407441beb213a"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
