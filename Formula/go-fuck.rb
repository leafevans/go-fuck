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
  version "2.23.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.23.0/go-fuck_2.23.0_darwin_arm64.tar.gz"
      sha256 "3d2bae3672a01923d0a8cbc0b5a27b7a430738315e24ac11544b24764b331d5f"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.23.0/go-fuck_2.23.0_darwin_amd64.tar.gz"
      sha256 "93e80598718ff1a0da956cfae281dfeb7c76d3b70614acf76acea2983aa96a49"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.23.0/go-fuck_2.23.0_linux_arm64.tar.gz"
      sha256 "30432ff68295177085129d56f3c5be57deafc760397730e9427cf83238e99d0e"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.23.0/go-fuck_2.23.0_linux_amd64.tar.gz"
      sha256 "b6069e4fa06eba4bfc43ba8e976c4e587978dd362ac4600fe4c4fb28bb553874"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
