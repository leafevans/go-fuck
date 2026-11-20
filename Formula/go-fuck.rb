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
  version "2.31.17"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.17/go-fuck_2.31.17_darwin_arm64.tar.gz"
      sha256 "040eb6b04148caa9d1ca0cfcd04c25d84d72b87a57c7a629d9619d6873579ac7"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.17/go-fuck_2.31.17_darwin_amd64.tar.gz"
      sha256 "f698ca6a3514546fd8236e0a1c07a1f5e2d2a66e3cef7067b02e156c5ed4603d"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.17/go-fuck_2.31.17_linux_arm64.tar.gz"
      sha256 "3392cbbf778e647b0a44c0cd23f16128e67c251b734e88580698af5557008f8b"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.17/go-fuck_2.31.17_linux_amd64.tar.gz"
      sha256 "59b6ba9812966a715d137ccd0ceba396d773116bd1a663db7b40ed16ac30e00f"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
