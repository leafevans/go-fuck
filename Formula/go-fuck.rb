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
  version "2.31.25"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.25/go-fuck_2.31.25_darwin_arm64.tar.gz"
      sha256 "05376712b70736a06e6e6bcc9e938e8727677a9b754ac30b8d200d423dda1d9d"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.25/go-fuck_2.31.25_darwin_amd64.tar.gz"
      sha256 "1ecb1133498a5d46e9743ad9f7fb9dae8b475bf008882a93c7ee7bea5b523e8b"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.25/go-fuck_2.31.25_linux_arm64.tar.gz"
      sha256 "87d781936d6ad46ace1f443b7f1242c6af702fab10936497cc8adb3be16ce487"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.25/go-fuck_2.31.25_linux_amd64.tar.gz"
      sha256 "677b86871b779a8e7fc9817737638f5991c3c40d7e4976113dae3192e962abb6"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
