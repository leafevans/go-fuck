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
  version "2.4.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.4.0/go-fuck_2.4.0_darwin_arm64.tar.gz"
      sha256 "0097ad223ad20ecddd1a08cbdc9f3c09a3e5fceed1af145c8282e9302ea96169"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.4.0/go-fuck_2.4.0_darwin_amd64.tar.gz"
      sha256 "cff7763b34d644f597d26605633f0296b970bb6555cebe92c7cb042ec52a0ca7"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.4.0/go-fuck_2.4.0_linux_arm64.tar.gz"
      sha256 "0d3598aeef277c701ab5596b786faf09483ab2ab138825ae9a3d67cfb8f7708a"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.4.0/go-fuck_2.4.0_linux_amd64.tar.gz"
      sha256 "1ab40fe7c6e8e2517eccc199b805de0988a2294ad25db05929e113d8921312a6"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
