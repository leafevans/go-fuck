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
  version "2.31.7"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.7/go-fuck_2.31.7_darwin_arm64.tar.gz"
      sha256 "6ec51a04596c735a928f575803852cb97971ab8d103efee1e9a95b69966c08bd"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.7/go-fuck_2.31.7_darwin_amd64.tar.gz"
      sha256 "86ead7cf12a3e42fc46095766f15e3947a7e39f8268bc104ec62cfb8328408c3"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.7/go-fuck_2.31.7_linux_arm64.tar.gz"
      sha256 "ef02c6aa759638f498079fecc01b91121f8f1684a04eb97a9d04ae65e6206b2b"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.7/go-fuck_2.31.7_linux_amd64.tar.gz"
      sha256 "0a90b969a034876ab2c40c61913f052731caec4f8dce3fe83bae7b70e5e607f9"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
