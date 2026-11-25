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
  version "4.2.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.2.1/go-fuck_4.2.1_darwin_arm64.tar.gz"
      sha256 "295dd0aa0e68e37a4d48e415fb00b1f79cf59484fb704f45cd818d2a40bec642"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.2.1/go-fuck_4.2.1_darwin_amd64.tar.gz"
      sha256 "79b1fa58f905c5d33742d567c5f8ac47640ec140c7ce7bd385844eeea6a5dad6"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.2.1/go-fuck_4.2.1_linux_arm64.tar.gz"
      sha256 "e834789b6d59b68785eb1e6190bfcb1f386b3c797e0249dcd8e0d85e95a47eb2"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.2.1/go-fuck_4.2.1_linux_amd64.tar.gz"
      sha256 "b5209246952042800f38e6fcca4a940c7421c3c5fc944e2085db93ee617c3c37"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
