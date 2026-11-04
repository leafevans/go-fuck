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
  version "1.2.2"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.2.2/go-fuck_1.2.2_darwin_arm64.tar.gz"
      sha256 "fedfea7df2d91a0e6344cb901a2f157ad7f714e3fbf333c0a8daf9ca1c7d9c22"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.2.2/go-fuck_1.2.2_darwin_amd64.tar.gz"
      sha256 "3793ee78c53fbdcb8ed6ac7ce70a06df78f60471f25f1a57ab092ef08bd9ee90"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.2.2/go-fuck_1.2.2_linux_arm64.tar.gz"
      sha256 "bb367136b80bc12ebf6c99db2fa18de743c4643aa00df128ea03ec38399a0bf0"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.2.2/go-fuck_1.2.2_linux_amd64.tar.gz"
      sha256 "096d5d545cbed00f00dc4846d208dc3379f5f0547db3d58db36d8e51e01fd605"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
