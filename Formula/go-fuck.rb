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
  version "1.7.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.7.1/go-fuck_1.7.1_darwin_arm64.tar.gz"
      sha256 "9c4c806110aa7340f8943b9eb11ece84d460e535e7dea57d79c11de4f9e04e5c"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.7.1/go-fuck_1.7.1_darwin_amd64.tar.gz"
      sha256 "63b0142eb58903f0ecd40ffe187cf7eb0b4da6c361e2f56d2e1a1afb0db61a16"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.7.1/go-fuck_1.7.1_linux_arm64.tar.gz"
      sha256 "9deb87551a3f1de72680373ce3a79b39fa6e7cc785fc6a09c1df65156b87bb04"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.7.1/go-fuck_1.7.1_linux_amd64.tar.gz"
      sha256 "84004faae6cd0398ad727663ad6829da7d584e2da5ebef81b883882553cf5bc3"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
