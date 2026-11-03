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
  version "0.11.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.11.0/go-fuck_0.11.0_darwin_arm64.tar.gz"
      sha256 "b27ffd0042d052295b23dc373183454a05b593cdbcb3c1e1ba4033245b1d4e4d"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.11.0/go-fuck_0.11.0_darwin_amd64.tar.gz"
      sha256 "c6dd04b6d390735fd858641194135122894f963ec190f2ce91367002e7fa27c4"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v0.11.0/go-fuck_0.11.0_linux_arm64.tar.gz"
      sha256 "9bc302700789a2bd3b3f1139241ed3fc146920bbb1bea118f57be1f8249cddc5"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v0.11.0/go-fuck_0.11.0_linux_amd64.tar.gz"
      sha256 "d58cd2400f26af64ba8559a1b909852ff8a610af67096c44a4162fec98df22c8"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
