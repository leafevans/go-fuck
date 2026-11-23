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
  version "3.0.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v3.0.0/go-fuck_3.0.0_darwin_arm64.tar.gz"
      sha256 "931b5b028d2f5bea5e95c9ccd92693256e604d8047d93162fb4911eb84d4610e"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v3.0.0/go-fuck_3.0.0_darwin_amd64.tar.gz"
      sha256 "e10bc3f031b9fd9be2e7408fb63c33eb0de75486595b1099651125ab9a3a41ab"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v3.0.0/go-fuck_3.0.0_linux_arm64.tar.gz"
      sha256 "86f5a94cf13efb89bfc8682068b8f7558e53e7aa0bf1c65db60378240bbc3a22"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v3.0.0/go-fuck_3.0.0_linux_amd64.tar.gz"
      sha256 "afbf1db95e20ef62408725e408a2b0e6d576568e5885168d3d98d19d14a7d219"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
