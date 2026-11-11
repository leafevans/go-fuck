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
  version "2.12.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.12.0/go-fuck_2.12.0_darwin_arm64.tar.gz"
      sha256 "3da20d8a5176de09b340ec0aae00b46bec02aa99a3c8e9a49db4a24c2d18c18e"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.12.0/go-fuck_2.12.0_darwin_amd64.tar.gz"
      sha256 "2bba29a9f17811ace30266a9fe09b65bbef0ade4f2774d9679b0c1444e3b8b99"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.12.0/go-fuck_2.12.0_linux_arm64.tar.gz"
      sha256 "65cf5be3b9c504081da1fabd50584d2f2dca875e89f47258bc408e33e8697512"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.12.0/go-fuck_2.12.0_linux_amd64.tar.gz"
      sha256 "129616f25b21481ae6e5af39ae0ecacbb227cbd74eda86f7acc629bd8d1c20f2"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
