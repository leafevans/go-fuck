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
  version "4.3.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.3.1/go-fuck_4.3.1_darwin_arm64.tar.gz"
      sha256 "0c5a2564b7a73ce2339fbcb0f512fbdcdb357fb13287e9572c8d7d30cc05fabd"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.3.1/go-fuck_4.3.1_darwin_amd64.tar.gz"
      sha256 "58d8ad3cbee0daf7690c6f8d0ef07d91d49b9e2646a3790885e63c1f3b637e5f"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.3.1/go-fuck_4.3.1_linux_arm64.tar.gz"
      sha256 "1368c1a376fc2658ba2f848df1e111eb213f4fdced873337a03627b0d40cf0e9"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.3.1/go-fuck_4.3.1_linux_amd64.tar.gz"
      sha256 "8e68eba63f92d880e065ec0a73f29dd68aed3d1ca3f3ca4b71d8de2646d9c101"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
