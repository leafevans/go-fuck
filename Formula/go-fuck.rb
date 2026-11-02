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
  version "0.7.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.7.0/go-fuck_0.7.0_darwin_arm64.tar.gz"
      sha256 "ac2c91aa160b9914764f7434a1ced38ff98e7db0713e2a1c59930610518a220e"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.7.0/go-fuck_0.7.0_darwin_amd64.tar.gz"
      sha256 "f17faf6dd7b10e614f77ed7528dfbe3465521676e3791773f739f538db70f0cc"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.7.0/go-fuck_0.7.0_linux_arm64.tar.gz"
      sha256 "e66efc72d78f44cf49ff87feb21fd0b02875d72873e72753a2db04c8ba53c0ff"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.7.0/go-fuck_0.7.0_linux_amd64.tar.gz"
      sha256 "1d8b9f9f46a9467acac2b2cbeb0786090cfd268b0df2b5c662c322e47bbe5453"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
