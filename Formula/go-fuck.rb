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
  version "2.19.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.19.0/go-fuck_2.19.0_darwin_arm64.tar.gz"
      sha256 "2331e66a976f194415051a55e2d20cebc3a8df2c99c8dd8641ad39e325455c71"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.19.0/go-fuck_2.19.0_darwin_amd64.tar.gz"
      sha256 "845f1151f378251c2fe60188c25ab09c55e2be17be142720b62edf0f35b2a379"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.19.0/go-fuck_2.19.0_linux_arm64.tar.gz"
      sha256 "4f990e0aa263613771b33aeb1a17cf8b29e79bc4111ed79725c426fd89ecfe3b"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.19.0/go-fuck_2.19.0_linux_amd64.tar.gz"
      sha256 "54763cc7ad2b7ed5c14b6c1a73f5a153242b66e3d204b2cb74d40706580834a9"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
