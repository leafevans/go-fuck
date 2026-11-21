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
  version "2.34.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.34.1/go-fuck_2.34.1_darwin_arm64.tar.gz"
      sha256 "6fd7d4fd3cddeae1702579c955591e47ce2fad7421520a59c8968cd59f80629d"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.34.1/go-fuck_2.34.1_darwin_amd64.tar.gz"
      sha256 "5e30db1c95260019bf612b237e55e6d237c7010188662720c418fd45b34863b4"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.34.1/go-fuck_2.34.1_linux_arm64.tar.gz"
      sha256 "c8854058372d89aacf43ae9da9c8f6cf5a57b33250f0458e0ad912d69a6d1a7f"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.34.1/go-fuck_2.34.1_linux_amd64.tar.gz"
      sha256 "93713806bf04988796c68e01f5dcf1d8fbf29a9d5a6423ea02d01dc980173409"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
