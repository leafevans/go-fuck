# Homebrew formula for go-fuck.
#
# This repository doubles as a Homebrew tap, so users can install the tool with:
#
#   brew tap leafevans/go-fuck https://github.com/leafevans/go-fuck
#   brew trust leafevans/go-fuck
#   brew install go-fuck
#
# The formula installs the prebuilt release archive for the current platform.
# Update it with "go run ./tools/update-packaging" when a new release is
# published (see packaging/README.md).
class GoFuck < Formula
  desc "Blazingly fast reimplementation of thefuck in Go"
  homepage "https://github.com/leafevans/go-fuck"
  version "0.2.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.2.0/go-fuck_0.2.0_darwin_arm64.tar.gz"
      sha256 "fedbb570ca6523202ae7ecd69135502ea6b8dac5dccca5fb66a51fe0c705d2b9"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.2.0/go-fuck_0.2.0_darwin_amd64.tar.gz"
      sha256 "af048defdeff5d38c77aa6a177c3a9ca5bd9db1666c7d69f5c3bb718160bafd0"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.2.0/go-fuck_0.2.0_linux_arm64.tar.gz"
      sha256 "b78e09b6eb27897dcd3ada05120c4c187069fc33b52065988eed602e45d462b6"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.2.0/go-fuck_0.2.0_linux_amd64.tar.gz"
      sha256 "29b14098ce7d886f78531ab4a61f534dd2d1c205347b0b68f6815e449a153776"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end