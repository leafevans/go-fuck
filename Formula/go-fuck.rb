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
  version "1.1.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.1.0/go-fuck_1.1.0_darwin_arm64.tar.gz"
      sha256 "25fda72aa499256167c7afe87080133cd3fa28a274097ffd70a09d0550e181b5"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.1.0/go-fuck_1.1.0_darwin_amd64.tar.gz"
      sha256 "ca2da24bc26cf54f4f9d9d62a53225144ba864ca91cb7c8792c9ca6ae3971907"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.1.0/go-fuck_1.1.0_linux_arm64.tar.gz"
      sha256 "7afd324d00983fb835a8b21f9224237387d3088be57b84b4a51fb8ee7532d93b"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.1.0/go-fuck_1.1.0_linux_amd64.tar.gz"
      sha256 "8f8e475f21cdaf19cf2ba9538de7f6ada202a290d3d6a2321e6aa180c96994a2"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
