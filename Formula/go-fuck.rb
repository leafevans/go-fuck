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
  version "2.11.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.11.0/go-fuck_2.11.0_darwin_arm64.tar.gz"
      sha256 "0f7cf5662528f90c9660bebe2e0f14103a72296a118cad04e843f9d9f48de14f"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.11.0/go-fuck_2.11.0_darwin_amd64.tar.gz"
      sha256 "3a5ed00cb140acf6cee4b1168cb93fcf89bccc20dd7c25299d2764336a61a2b1"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.11.0/go-fuck_2.11.0_linux_arm64.tar.gz"
      sha256 "61e81794e9eef2caeb88ff86df36c3e11559a698a0df194c02b26d463dac0786"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.11.0/go-fuck_2.11.0_linux_amd64.tar.gz"
      sha256 "d6e3513d9be6e2852e889e1217f315bfa0b0cce24ac5cf1a93df18d91770c9bf"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
