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
  version "1.11.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.11.0/go-fuck_1.11.0_darwin_arm64.tar.gz"
      sha256 "33a7240191d11a124bbd181bf548c523fbda7942f931e7f2d4e33c904abd1f7c"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.11.0/go-fuck_1.11.0_darwin_amd64.tar.gz"
      sha256 "8c35aebcdd0ed5fc5038f1ea3cb22e5969d8986b55137fd08de8e25e1f91bd70"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.11.0/go-fuck_1.11.0_linux_arm64.tar.gz"
      sha256 "be0816cb13ccd53993c8235d9977eda02933a4cd6552339c6a837e655c1198e0"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.11.0/go-fuck_1.11.0_linux_amd64.tar.gz"
      sha256 "bb2b77182e3f1f0924449b516a7567dbc1b4c4ef601e8e1141acd8ce20ce3d45"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
