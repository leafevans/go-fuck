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
  version "2.12.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.12.1/go-fuck_2.12.1_darwin_arm64.tar.gz"
      sha256 "3704da97494a8e21fd5a26287f408dfbcc4ed6cf1b1a948739f3df1a3d9c9ba0"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.12.1/go-fuck_2.12.1_darwin_amd64.tar.gz"
      sha256 "8c6f20ac8b3638f43b489998c48473c6a798c34bf13c7089dea0795860d6516f"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.12.1/go-fuck_2.12.1_linux_arm64.tar.gz"
      sha256 "bf762dd0c8a1c7c1fc139957012d82bb9ccf66a8aa714f3da0dcf92a3a042ccc"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.12.1/go-fuck_2.12.1_linux_amd64.tar.gz"
      sha256 "647abd152a896eff7b094b511531cf06dbbd3851bb72e05cc0fddc39c5997799"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
