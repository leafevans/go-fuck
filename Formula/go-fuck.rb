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
  version "2.3.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.3.0/go-fuck_2.3.0_darwin_arm64.tar.gz"
      sha256 "580927bed47100a65ef46d4ee360408ac192ff7cf1c1e532aee04cda90754aa3"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.3.0/go-fuck_2.3.0_darwin_amd64.tar.gz"
      sha256 "1d74cf929bcf2d684a217e6fc21be4b1afaea001bd86da9527805d308a4cda67"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.3.0/go-fuck_2.3.0_linux_arm64.tar.gz"
      sha256 "1d69f55de5c70609c6756068139297f52a015ebf512c3f7c4cea48e3f7176c7d"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.3.0/go-fuck_2.3.0_linux_amd64.tar.gz"
      sha256 "131b14714f33f423f19dd66829d92a34aafc79d0e21d2e44693082b923257a08"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
