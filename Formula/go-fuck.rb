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
  version "2.20.6"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.6/go-fuck_2.20.6_darwin_arm64.tar.gz"
      sha256 "5f7e1200593092080f7e4879977e346101881ad41cabfc0cef08afdbeb924beb"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.6/go-fuck_2.20.6_darwin_amd64.tar.gz"
      sha256 "f6fed93bac308a544498ed55ab954008d952a03791ad1acbd20d109119ef8093"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.6/go-fuck_2.20.6_linux_arm64.tar.gz"
      sha256 "80388e358f139a9849a3eb6b8d6f57091f55cf38aaae089c7ebc99dc70784dbd"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.6/go-fuck_2.20.6_linux_amd64.tar.gz"
      sha256 "f01dded4e73ff304c1ee11bb70d651c8cf1196c27db3770361f15fa751510033"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
