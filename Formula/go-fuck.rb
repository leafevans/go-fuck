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
  version "4.2.2"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.2.2/go-fuck_4.2.2_darwin_arm64.tar.gz"
      sha256 "e40cf313b1f33348900d65a6a9f3efb9ab437488b93a6db89f0539b0dafec981"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.2.2/go-fuck_4.2.2_darwin_amd64.tar.gz"
      sha256 "ff8f1afa753381db8af27abf5d25b2a60c80e4db838676f01184415fe141176f"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.2.2/go-fuck_4.2.2_linux_arm64.tar.gz"
      sha256 "75361c70d0f6c17dfb4bdc0fade26fbe2cd3b3cb8d684479fa42f0bab1ac8c8a"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.2.2/go-fuck_4.2.2_linux_amd64.tar.gz"
      sha256 "785d6405fa337c87d6b9c78ad15de7fa14b6da3f9a182b78f80e6916342ca5dc"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
