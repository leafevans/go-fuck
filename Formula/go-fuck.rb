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
  version "2.31.13"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.13/go-fuck_2.31.13_darwin_arm64.tar.gz"
      sha256 "57fe17c2b8fc87a9318bff5b336f3607813ec87f602b3eaa7ae7026e7a708aa6"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.13/go-fuck_2.31.13_darwin_amd64.tar.gz"
      sha256 "690ea45eca71c727bf1959f9dbc24663b6827064821e99b5bd5ac99b3ef5691b"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.13/go-fuck_2.31.13_linux_arm64.tar.gz"
      sha256 "f23c3fcfd21a90aad7cc1fd41b9de60cb51e891698783e2ffc04a74c7ffae5b8"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.13/go-fuck_2.31.13_linux_amd64.tar.gz"
      sha256 "5918158d9b0f1af37e160053b973f9ea894ea258a1c4335e6ea6de433f5b7d6a"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
