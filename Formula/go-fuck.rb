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
  version "2.31.8"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.8/go-fuck_2.31.8_darwin_arm64.tar.gz"
      sha256 "4029bb5a39262becd2f79d8d61f1a312f7e40ae9ac7b77d972877f43842e0afe"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.8/go-fuck_2.31.8_darwin_amd64.tar.gz"
      sha256 "2d6ad8e0955f2d554aa5767255d5b39770f4df5b870b58a722af625798b4a40d"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.8/go-fuck_2.31.8_linux_arm64.tar.gz"
      sha256 "9996db5b8f48efeb10db0f2120ac7045a2fdcdf53c4357a732de7ef3be453502"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.8/go-fuck_2.31.8_linux_amd64.tar.gz"
      sha256 "184328e3339983a75c730015cb4e94751a9c2c0440368aa1e0efe97d979619a1"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
