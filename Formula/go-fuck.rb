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
  version "1.4.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.4.1/go-fuck_1.4.1_darwin_arm64.tar.gz"
      sha256 "e75f87265afd38c7b88d1bab5c15497a8c1303c1c9f7d0dc4c5f3e918caea399"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.4.1/go-fuck_1.4.1_darwin_amd64.tar.gz"
      sha256 "8399a6fe6628c9fb483427b107dbfb04933ea90a559257288944122ce4c9fd82"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.4.1/go-fuck_1.4.1_linux_arm64.tar.gz"
      sha256 "0a32207ffab238038c3a3e69f99b5ac55cd97c9e0f1f868313506be15d60405e"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.4.1/go-fuck_1.4.1_linux_amd64.tar.gz"
      sha256 "e458cec4c21ccaa8cc602e10169a5b8c8b72f372e966cfc332277ca93f0f4737"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
