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
  version "1.6.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.6.0/go-fuck_1.6.0_darwin_arm64.tar.gz"
      sha256 "7085c4e50c5a7a81476f221c9558beb029fdbb99edeccb322bd50f75e36f3ddb"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.6.0/go-fuck_1.6.0_darwin_amd64.tar.gz"
      sha256 "cdca2cf50e130435f775e834748c00e11c07dea01d7074ef397849d1b94705e5"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.6.0/go-fuck_1.6.0_linux_arm64.tar.gz"
      sha256 "63667071a9882a11a409ba31bc81a94ba60b765e48f9475b7d670063be8bd266"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.6.0/go-fuck_1.6.0_linux_amd64.tar.gz"
      sha256 "43931c570897ac3b0c12ad2634ab402a9b3945972d0cba4f0b6bd7262a2fe8e6"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
