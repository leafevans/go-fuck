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
  version "2.24.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.24.0/go-fuck_2.24.0_darwin_arm64.tar.gz"
      sha256 "fc2aa3b4705b97087ee54636b6203b5eb5365f48a343208ec1f60a8a9bbcfa9d"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.24.0/go-fuck_2.24.0_darwin_amd64.tar.gz"
      sha256 "21dbd469c18b35671e1836ffc5568a3a4f5e36579c250832c8e03b94a996e00e"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.24.0/go-fuck_2.24.0_linux_arm64.tar.gz"
      sha256 "dcaeb83099fc3cd835837e464e6b92141d226bed026be65d37740fad4f0e2b48"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.24.0/go-fuck_2.24.0_linux_amd64.tar.gz"
      sha256 "743c695241986676397db840976e229ed22b30dab7233f617c45258422a7a8e2"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
