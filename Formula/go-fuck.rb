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
  version "2.8.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.8.0/go-fuck_2.8.0_darwin_arm64.tar.gz"
      sha256 "bac376e1d344ba88dc1410c27eaf525611d4ac2ff6c1869d20d990988ca5ce0d"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.8.0/go-fuck_2.8.0_darwin_amd64.tar.gz"
      sha256 "a0fd5555da22e2178c54a6e73c8d76406b24929a967fbac0f221dd355ff26253"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.8.0/go-fuck_2.8.0_linux_arm64.tar.gz"
      sha256 "92632ec441a7939aa7567852c819e5e51e01b031763c3aa30bec6c5b68d9c9eb"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.8.0/go-fuck_2.8.0_linux_amd64.tar.gz"
      sha256 "5676bfb2e9a162ce0c624eea3ac3c4937787c2a2b551efb8ed6f1c32f4471a2c"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
