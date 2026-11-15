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
  version "2.30.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.30.0/go-fuck_2.30.0_darwin_arm64.tar.gz"
      sha256 "6a7db61c4b18f40509677fd8cc5c930b7c2186b87ba30fe4a588d37cceb92ef8"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.30.0/go-fuck_2.30.0_darwin_amd64.tar.gz"
      sha256 "ff2ee6fa5c03860565370ac456e6d4211f36b4790d010fdb910ebd785c7732ed"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.30.0/go-fuck_2.30.0_linux_arm64.tar.gz"
      sha256 "958a581ddd1d5742a72e7f114c65a7facaf78715999b6742a5387912a296753f"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.30.0/go-fuck_2.30.0_linux_amd64.tar.gz"
      sha256 "112f1fe4554067cde5093b4ad5f66d4f2c7ffaf929e5c21546f8a3305a2b342c"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
