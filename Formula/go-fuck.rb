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
  version "4.0.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.0.1/go-fuck_4.0.1_darwin_arm64.tar.gz"
      sha256 "17adadce6c4dcd90f40ac7214b536ec60baf7562bddcd67132564217beb7b516"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.0.1/go-fuck_4.0.1_darwin_amd64.tar.gz"
      sha256 "fd5efb0bacb3be8ee570e53bc625f456fef98b1f54ad6aae9f4f22bcd69996ac"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.0.1/go-fuck_4.0.1_linux_arm64.tar.gz"
      sha256 "72753f6a034e5a1c00cae80c7156acabfe36fecd7c5a0775410a4f24a20cb9e3"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.0.1/go-fuck_4.0.1_linux_amd64.tar.gz"
      sha256 "591807cb3b58143b211f336c9344ebfbe56cbc421cc602922a61ce2a3ead6fa2"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
