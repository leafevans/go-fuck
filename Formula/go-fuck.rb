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
  version "2.15.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.15.1/go-fuck_2.15.1_darwin_arm64.tar.gz"
      sha256 "1cc0ebd246bb83e4f9875fc3884c90645ce3ebba054406a438bec08949556182"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.15.1/go-fuck_2.15.1_darwin_amd64.tar.gz"
      sha256 "79cdbe1d3f6f7a2718f39fbf280e01dd5f0ce6fc158efe8df3023870776479b6"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.15.1/go-fuck_2.15.1_linux_arm64.tar.gz"
      sha256 "4e9b8257713bfc2dbf5381212b3a49587a38155f80332e43b4591c781ea3022d"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.15.1/go-fuck_2.15.1_linux_amd64.tar.gz"
      sha256 "007d1147f5d76ab9e468d0cdadc696916637df02e8af1454df93bb744ad7e010"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
