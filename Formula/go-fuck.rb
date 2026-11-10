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
  version "1.10.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.10.0/go-fuck_1.10.0_darwin_arm64.tar.gz"
      sha256 "09a7bfeccfce762329141078eabb16f577bd5d9dfb803a5aeee1dc0c8be319e1"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.10.0/go-fuck_1.10.0_darwin_amd64.tar.gz"
      sha256 "d302e1c1afd94f5a9780de93aa75112efe2e5040f89c55849138d16f847f9be2"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.10.0/go-fuck_1.10.0_linux_arm64.tar.gz"
      sha256 "cc045b663b63fffe5a024ef8fe36e053e131a4c6b5a5f2fa9c18fcb306ed3e12"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.10.0/go-fuck_1.10.0_linux_amd64.tar.gz"
      sha256 "3b9ecf1bd42b382793b27ec4f346daeca73a38f6bff76e1f3395bf5784c26d0a"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
