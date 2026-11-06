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
  version "1.3.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.3.1/go-fuck_1.3.1_darwin_arm64.tar.gz"
      sha256 "50fd35d9c2c896c946978e9a46b6a74501c354e7695533bed906704f709ccc7f"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.3.1/go-fuck_1.3.1_darwin_amd64.tar.gz"
      sha256 "ec2f646cc1efd13795cdb0cbce1aeba152d7b49b35eaf25561b67f57c65758cb"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.3.1/go-fuck_1.3.1_linux_arm64.tar.gz"
      sha256 "8edb17cf470117e0fac935c1d52f95fa390bc01d2cdad7516ba8bd1f61a192be"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.3.1/go-fuck_1.3.1_linux_amd64.tar.gz"
      sha256 "6b18a17b3becb2825f9f6acb963efab4d02367f35c68b0fcbb5abbb28c1d9ea6"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
