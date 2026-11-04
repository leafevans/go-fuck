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
  version "1.0.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.0.0/go-fuck_1.0.0_darwin_arm64.tar.gz"
      sha256 "8a45fc8501c0887b2179d4e8450be199c7aa1183c4ac7696b0af7b9b829aaa45"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.0.0/go-fuck_1.0.0_darwin_amd64.tar.gz"
      sha256 "ce74a740a0169089cb517eadc49c51aa6ab6a206235f225e30c5e9131dcdf51e"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.0.0/go-fuck_1.0.0_linux_arm64.tar.gz"
      sha256 "721da4ca61b4ffc066aece55363839537764059944ba0501c60622306737cdc2"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.0.0/go-fuck_1.0.0_linux_amd64.tar.gz"
      sha256 "200b8927ffdea32d6a9865e76bd2f2bdb1a3e9e1a50db8527b88b0e5417c781b"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
