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
  version "2.16.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.16.0/go-fuck_2.16.0_darwin_arm64.tar.gz"
      sha256 "f809c5c04eb6394a069562bb1108845c1dba66c32133ff51fb0482aa58d66121"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.16.0/go-fuck_2.16.0_darwin_amd64.tar.gz"
      sha256 "6a1e77e622a734e6b4f7dc9f2d768da30bccaf0efd8eb97e5f2a6d7a6b7b0116"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.16.0/go-fuck_2.16.0_linux_arm64.tar.gz"
      sha256 "577a7a7211e28a42081ccad7e1735b8568bb74660336cdf6eaed30e037d14677"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.16.0/go-fuck_2.16.0_linux_amd64.tar.gz"
      sha256 "e96b827bd56db3f5c69566d221de99a5201ecc8dfc604148954ee476e5b43b35"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
