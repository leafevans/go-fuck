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
  version "4.2.4"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.2.4/go-fuck_4.2.4_darwin_arm64.tar.gz"
      sha256 "f918c27d49c6a6a9d74fcb8a77d7bc861bf69963c6b238dcc9b57e690e6def71"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.2.4/go-fuck_4.2.4_darwin_amd64.tar.gz"
      sha256 "dfbae503a9091482e2913faeb3e6768e4d9a8455caee6cb4e8546fb4462aa0a4"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.2.4/go-fuck_4.2.4_linux_arm64.tar.gz"
      sha256 "644d9de961936a35cd2633f2d0ff757ecbf9390956b0704e445cf63b2b23e197"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.2.4/go-fuck_4.2.4_linux_amd64.tar.gz"
      sha256 "975d57ed83157d193fcebf4eee6299ec9e19b07cf081fa5ea32fdec434a85094"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
