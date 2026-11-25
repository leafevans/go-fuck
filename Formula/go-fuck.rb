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
  version "4.0.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.0.0/go-fuck_4.0.0_darwin_arm64.tar.gz"
      sha256 "c2979f2d2fcdd5b67ca7b5704dd69735619ab465a2b3be0e4abb5d5aed1e695c"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.0.0/go-fuck_4.0.0_darwin_amd64.tar.gz"
      sha256 "38bae8ea1afb80bb2b529bdd510943568dcbdb7cb7d4a78ac35885f1232e5120"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.0.0/go-fuck_4.0.0_linux_arm64.tar.gz"
      sha256 "1055f75ae1d4e50e499cb220211e9a28a52e7ad776d18f58ae45aef81d617e30"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.0.0/go-fuck_4.0.0_linux_amd64.tar.gz"
      sha256 "ce3c4160b71b48204cbf21aaf7782ce959f85502b3a5ceb14a9e35dae1de7fad"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
