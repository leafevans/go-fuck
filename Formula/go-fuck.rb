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
  version "2.31.24"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.24/go-fuck_2.31.24_darwin_arm64.tar.gz"
      sha256 "5ae48efa2177069efbeab731fd083a9ccde26a3c42d996abf53262fe361de41b"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.24/go-fuck_2.31.24_darwin_amd64.tar.gz"
      sha256 "e0155d8fc34dce341e2d7d57fb3c39d11de764be30c3b1ab770686c56199b17c"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.24/go-fuck_2.31.24_linux_arm64.tar.gz"
      sha256 "a1010378bcc34ebd7f15b9b8172b2069f9c4ea6b478afdd88044d26ff5babcab"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.31.24/go-fuck_2.31.24_linux_amd64.tar.gz"
      sha256 "ac74a466b9020d1977dfbbc1339eb2bfb06e3cfd3262e1b55770fea6d49b6611"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
