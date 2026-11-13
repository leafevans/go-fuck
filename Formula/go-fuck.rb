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
  version "2.27.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.27.0/go-fuck_2.27.0_darwin_arm64.tar.gz"
      sha256 "0f949d7b9389bf60305f08096dc37dba12e59aa872e84e5f8bb132ec3627d5e3"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.27.0/go-fuck_2.27.0_darwin_amd64.tar.gz"
      sha256 "9d80e3764a61b57000a1e01d093254105167d15d0c8595c8582843ba56eff836"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.27.0/go-fuck_2.27.0_linux_arm64.tar.gz"
      sha256 "4c694d31c432777fa1993a1c12f53cc93d10a000464d58c59762607cbe8f4664"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.27.0/go-fuck_2.27.0_linux_amd64.tar.gz"
      sha256 "ec6ca38d0cbc95c9b4727806c72ad28eaf0908efa5323e29b469d76be18cb363"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
