# Homebrew formula for go-fuck.
#
# This repository doubles as a Homebrew tap, so users can install the tool with:
#
#   brew tap leafevans/go-fuck https://github.com/leafevans/go-fuck
#   brew install go-fuck
#
# The formula installs the prebuilt release archive for the current platform.
# Update the version and the four checksums when a new release is published
# (see packaging/README.md).
class GoFuck < Formula
  desc "Blazingly fast reimplementation of thefuck in Go"
  homepage "https://github.com/leafevans/go-fuck"
  version "0.1.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v#{version}/go-fuck_#{version}_darwin_arm64.tar.gz"
      sha256 "6ae4601d94a7c1434bd00bbc3cc7a1150b844c60f693da45d62a1f515de584ee"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v#{version}/go-fuck_#{version}_darwin_amd64.tar.gz"
      sha256 "1c2d4c29fa1efebec51c0953231c998d2b1493d1bab665a6b00ebcea6c44da96"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v#{version}/go-fuck_#{version}_linux_arm64.tar.gz"
      sha256 "c819e389196cfb34b7fba3937099b2ced6b70b5388953c1092536d58e3e7e4df"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v#{version}/go-fuck_#{version}_linux_amd64.tar.gz"
      sha256 "8f3065a5c9852ba2aa52cc072c48597d37b042bdb46852579675213d12ff0799"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end