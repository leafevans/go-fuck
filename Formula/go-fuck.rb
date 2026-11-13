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
  version "2.25.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.25.0/go-fuck_2.25.0_darwin_arm64.tar.gz"
      sha256 "a3834ef3c2a64829d22a04114ae4f585d66d58abeab10b263c3fb5597feb735d"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.25.0/go-fuck_2.25.0_darwin_amd64.tar.gz"
      sha256 "f5e1f3014f3ee1bce60f55330fcf464bc0a017305343f5e26ee5bba02e4c7747"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.25.0/go-fuck_2.25.0_linux_arm64.tar.gz"
      sha256 "dc71b896e1ba10a2cbae9ace0e687f1bf0ae102b170554a3ef4dd5b7ac31f1bb"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.25.0/go-fuck_2.25.0_linux_amd64.tar.gz"
      sha256 "dcb50e4521371f0936a808d6f2fe9efad38806299cbc5cfe296712d92da79fb3"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
