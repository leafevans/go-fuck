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
  version "1.2.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.2.1/go-fuck_1.2.1_darwin_arm64.tar.gz"
      sha256 "f57687eda4f25b30497994fc2561113a90d03a387f70e70a27530cb1eeaf6b47"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.2.1/go-fuck_1.2.1_darwin_amd64.tar.gz"
      sha256 "dc099835cc369e4320619a2d5a419e9a80b263c0a15bd1ba8444d1ae27598866"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.2.1/go-fuck_1.2.1_linux_arm64.tar.gz"
      sha256 "b0e9fc9e24bbdf57f4e22e5fac95b20358b6b9587e96387f2c4764af861f67f2"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.2.1/go-fuck_1.2.1_linux_amd64.tar.gz"
      sha256 "ba1c6881e8eed38c9bb3e1fd8a4677e6fde08a88bb98168c7074152c68b7423f"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
