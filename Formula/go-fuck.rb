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
  version "0.2.3"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.2.3/go-fuck_0.2.3_darwin_arm64.tar.gz"
      sha256 "18bd0efcd94b114441462155fc69945564f992e879904d1ffd1d412bc66987bb"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.2.3/go-fuck_0.2.3_darwin_amd64.tar.gz"
      sha256 "cad6ee6866e62507dd6e644fa33baec81b6df8114ae6337e2061ee182d562077"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.2.3/go-fuck_0.2.3_linux_arm64.tar.gz"
      sha256 "a2732a67fa6878409fa11cd64cb151822ac6686cfd20bc9f7ea5576b8487e2f5"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.2.3/go-fuck_0.2.3_linux_amd64.tar.gz"
      sha256 "9954bc4101dd6af6ef07f93ddac87fc422300831542c63fae25def0e46d402e2"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
