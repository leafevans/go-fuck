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
  version "2.31.26"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.26/go-fuck_2.31.26_darwin_arm64.tar.gz"
      sha256 "4f5363be26f788e09514569bf22c951d93918994fb7692cd3d2b17662e1d52be"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.26/go-fuck_2.31.26_darwin_amd64.tar.gz"
      sha256 "06ea15c38ac3126a7e91dda26f7b8ba903cf9d548fa7fde281e7b9fb33778841"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.26/go-fuck_2.31.26_linux_arm64.tar.gz"
      sha256 "4af74e1070e4f4dab8130930f8e29f75b284653d0a1e2efbf6f859be0409def9"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.26/go-fuck_2.31.26_linux_amd64.tar.gz"
      sha256 "1335e4092a91a3d98c6412c101d366d8588a7ff58e37be48890fe6b4d7a26d36"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
