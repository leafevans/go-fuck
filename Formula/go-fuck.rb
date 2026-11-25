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
  version "4.1.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.1.0/go-fuck_4.1.0_darwin_arm64.tar.gz"
      sha256 "07b61acecd2312a85c37f1786ccffc1ca1c3976ccad8aaa0a3148ea41c61a465"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.1.0/go-fuck_4.1.0_darwin_amd64.tar.gz"
      sha256 "46919cc8a511746fba1508af9d922f05a8a38872fa8fbe55edd52978cf8f8b42"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.1.0/go-fuck_4.1.0_linux_arm64.tar.gz"
      sha256 "37dda8183ab62018ae0f034a321f1d173c6e7ba2fef91d8a5f77c7024ec31241"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.1.0/go-fuck_4.1.0_linux_amd64.tar.gz"
      sha256 "157233933ae4e551da4b7ccb6ae6d51be5c2bf7ededf2d94a61fbb6b9b819433"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
