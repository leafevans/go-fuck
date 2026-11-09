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
  version "1.7.2"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.7.2/go-fuck_1.7.2_darwin_arm64.tar.gz"
      sha256 "13b68e4f135e41e947aa835983fe58f043f3b134143a9d70b8467ef4b05169fc"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.7.2/go-fuck_1.7.2_darwin_amd64.tar.gz"
      sha256 "35a40ce7552ecaafa957f887b57b8d1379ce357cb882cf185ea79a1ddc92301e"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.7.2/go-fuck_1.7.2_linux_arm64.tar.gz"
      sha256 "76dda749071f9a5ff4cf9c135caef8f21590b290efd7ed7cfa5cca8834437255"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.7.2/go-fuck_1.7.2_linux_amd64.tar.gz"
      sha256 "a5a3363090c21ae9400bd60c5d919c86ff22a1bf20ae4a807d73edccc2eb8932"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
