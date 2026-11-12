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
  version "2.20.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.0/go-fuck_2.20.0_darwin_arm64.tar.gz"
      sha256 "82401b563d9169546ac29ca6ba0a425156f7e766d0b9d8d3cd50f53142f14b61"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.0/go-fuck_2.20.0_darwin_amd64.tar.gz"
      sha256 "ae6e3f35954343beab6adc8aa0d8d8b591dca83bb610eeb5d4050d2cc6298ddb"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.0/go-fuck_2.20.0_linux_arm64.tar.gz"
      sha256 "cea2456b8c2e692f706bb7f7d0aa44990cdf23004d6965a2c69f83aeac45fe7b"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.0/go-fuck_2.20.0_linux_amd64.tar.gz"
      sha256 "2f84a52c98f20629f5d3fcefe1f6f941b48acfc80e5f5c80a2d55b08ceebf5c4"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
