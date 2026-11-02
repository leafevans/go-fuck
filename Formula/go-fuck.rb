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
  version "0.5.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.5.0/go-fuck_0.5.0_darwin_arm64.tar.gz"
      sha256 "f0612ef4b6136442e977747453928bcc66eb4e6796fb39371e472556064ae3ad"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.5.0/go-fuck_0.5.0_darwin_amd64.tar.gz"
      sha256 "0b88418d60fc21c64e632043e915d18cc344be4cf994ee7d1721267657f3ce0f"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.5.0/go-fuck_0.5.0_linux_arm64.tar.gz"
      sha256 "6bcf45ade0745e31e67b32eac40a2f38a29895b9e681d37be3759c12366f7e47"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.5.0/go-fuck_0.5.0_linux_amd64.tar.gz"
      sha256 "ef36b3df11b9b68a517f20fe06517c0b6d94f625a44e14141752cd7c06cdaa74"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
