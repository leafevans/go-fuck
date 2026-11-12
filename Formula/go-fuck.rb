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
  version "2.14.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.14.1/go-fuck_2.14.1_darwin_arm64.tar.gz"
      sha256 "acad072555d159505d5d882d7a9b40f5a98d54657fdbc7c03b708b52b5fee8cf"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.14.1/go-fuck_2.14.1_darwin_amd64.tar.gz"
      sha256 "54bd20f2679dabf2b0bf733afd6f746e029342594016fa75d20ffa18331d01c8"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.14.1/go-fuck_2.14.1_linux_arm64.tar.gz"
      sha256 "68731a2485971b6295756a44b92f5c5aa3d1d289f8965c1bc5414a18cfa5d1d9"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.14.1/go-fuck_2.14.1_linux_amd64.tar.gz"
      sha256 "4d212429fbac0b20002dbe0240a34be30dc3f486a59eb6ad8ca5c2d8ccd6403a"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
