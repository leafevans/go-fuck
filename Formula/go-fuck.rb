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
  version "1.8.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.8.0/go-fuck_1.8.0_darwin_arm64.tar.gz"
      sha256 "05bda4925cc4d4559dde9ecb487d180c8c4af8fc9cb43e6edd5b2b496903af80"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.8.0/go-fuck_1.8.0_darwin_amd64.tar.gz"
      sha256 "11167b0794057b46d3f1463d24d982a948f175dd7a4102c5ccf178ccd42b9bbc"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.8.0/go-fuck_1.8.0_linux_arm64.tar.gz"
      sha256 "d75a35b4ff1d6bbf27651a7f63be2767d53b9cd0bb45e497eb5f3ee9e19c0356"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.8.0/go-fuck_1.8.0_linux_amd64.tar.gz"
      sha256 "46d6dad03f4aa4dc9b17f246959f6792aa133346f219a61a1a51f7957a1d56c4"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
