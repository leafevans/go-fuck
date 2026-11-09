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
  version "1.7.3"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.7.3/go-fuck_1.7.3_darwin_arm64.tar.gz"
      sha256 "a9a4698ff6bf5670e6471fc5b72352c4ae4ba4cc742678789e2b3945c18a3d26"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.7.3/go-fuck_1.7.3_darwin_amd64.tar.gz"
      sha256 "96775e5d8e215f895aa22cd7ee6665d35dc4cdcff686ca77aa73c7b29a19612b"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v1.7.3/go-fuck_1.7.3_linux_arm64.tar.gz"
      sha256 "f327cd5a0559a4ff2fb7b304e64e1d36468a57a51f629f64a469fbeef686aba7"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v1.7.3/go-fuck_1.7.3_linux_amd64.tar.gz"
      sha256 "6de121ec6d88ab36e3aa91849c74068275d4e9be952de7a6a314d189ca87d3ad"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
