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
  version "2.31.4"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.4/go-fuck_2.31.4_darwin_arm64.tar.gz"
      sha256 "bcfff44c7a31afaac605ec72bf063f2c16a785303f0952caf0a2c02d5168fbf4"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.4/go-fuck_2.31.4_darwin_amd64.tar.gz"
      sha256 "2977db087c8e6a241e1f6f559098ff187261de80324c5034563e5e151a18db50"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.4/go-fuck_2.31.4_linux_arm64.tar.gz"
      sha256 "3720706226408c73362dd1934e93f4c497517ad250a9de667098c2edb5c5fe1d"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.4/go-fuck_2.31.4_linux_amd64.tar.gz"
      sha256 "3c75a13fedb5179015619bdd23cf17844989ad662510fdf5b15c9cbc44e74536"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
