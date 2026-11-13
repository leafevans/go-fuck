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
  version "2.20.7"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.7/go-fuck_2.20.7_darwin_arm64.tar.gz"
      sha256 "4c79e97aa9d69b88d607f05471447c320154bbc6e00123c93c7d0306a5e296b3"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.7/go-fuck_2.20.7_darwin_amd64.tar.gz"
      sha256 "57ad9cad9967541fc3bc6bb316157f2b3811dd29798883c9806e9addccb03a70"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.7/go-fuck_2.20.7_linux_arm64.tar.gz"
      sha256 "6b9b5210dc5a90157603b2d9063c91a30c05d79d94e75454056ecb4a08f1b797"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.7/go-fuck_2.20.7_linux_amd64.tar.gz"
      sha256 "ea7f2d7600ef01a02f8de5612ac661c2374da9c4c789f0e9fff38768d9745b93"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
