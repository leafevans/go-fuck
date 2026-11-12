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
  version "2.20.3"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.3/go-fuck_2.20.3_darwin_arm64.tar.gz"
      sha256 "60607e77b74fca5867dfa47aa3d83b39b82d107ab7110e23bee15f83371c9092"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.3/go-fuck_2.20.3_darwin_amd64.tar.gz"
      sha256 "5f4c256e8c45c5b03761d1ba39842e54f54fcfcbc1fc7326d2ac8fb728b979d7"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.3/go-fuck_2.20.3_linux_arm64.tar.gz"
      sha256 "4b432848e1b2184984e4c59927da46e3f099cdfe855f6514fd3ea1d360c9909d"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.3/go-fuck_2.20.3_linux_amd64.tar.gz"
      sha256 "e7f6486caab6c707572eb93e45f2093974665aa5c4058c5c6d1e081f352004f6"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
