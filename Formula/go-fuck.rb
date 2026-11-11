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
  version "2.10.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.10.0/go-fuck_2.10.0_darwin_arm64.tar.gz"
      sha256 "9b332d91701c2fe9cc520cbf112736f2ac79ccb42ff901a851e8a743cbb9bf19"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.10.0/go-fuck_2.10.0_darwin_amd64.tar.gz"
      sha256 "ea1a7649b9a6564eff31fa5524ae003864c112a25fdf3aa88fb28d3598b2728f"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.10.0/go-fuck_2.10.0_linux_arm64.tar.gz"
      sha256 "6262654a5469e4fb18087e9c3f03654af46e25cf7d544dafdaab465c45acd250"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.10.0/go-fuck_2.10.0_linux_amd64.tar.gz"
      sha256 "0ab0274d05601ae2a54c1f4f3a1c1e3dc56b3c0e5e68dc6a5ef5475b916cf19b"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
