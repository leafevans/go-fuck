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
  version "2.31.12"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.12/go-fuck_2.31.12_darwin_arm64.tar.gz"
      sha256 "ff04c90596121cc71ae721c755171f74f622dd09826812eeb0ab29f2b66ecdb7"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.12/go-fuck_2.31.12_darwin_amd64.tar.gz"
      sha256 "78da4556894a2670e4fec0878d16de311bb5a5d166c70845d51edb4c58bcf4ac"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.12/go-fuck_2.31.12_linux_arm64.tar.gz"
      sha256 "bdd97fe7a900fd00e1098592697c14057c9afc8be96555388075d84c4d9d809a"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.12/go-fuck_2.31.12_linux_amd64.tar.gz"
      sha256 "cf5332b5887c7967bec5c11401c86dc0030bbad12dee51b1084eff025f709899"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
