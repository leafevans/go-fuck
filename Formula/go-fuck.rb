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
  version "2.1.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.1.0/go-fuck_2.1.0_darwin_arm64.tar.gz"
      sha256 "f4e5a580c74a6eb7cae289b4d966f432ea99a574fb2aa53b69a9657d54c779f0"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.1.0/go-fuck_2.1.0_darwin_amd64.tar.gz"
      sha256 "9bb06ea938b8a5ded6d6b74a525560e0dd96bf6ac9bbb664d5185a2f00bff835"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.1.0/go-fuck_2.1.0_linux_arm64.tar.gz"
      sha256 "fab7f057c3e222bc0f31b26bb92fc7ba8f710288315a3cdeef98c5da4bbbd408"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.1.0/go-fuck_2.1.0_linux_amd64.tar.gz"
      sha256 "dfab7bdc61c2eb201ec8e19ab451c6431480018df8314b5407ca804a7a514819"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
