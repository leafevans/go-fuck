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
  version "4.2.3"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.2.3/go-fuck_4.2.3_darwin_arm64.tar.gz"
      sha256 "870440d25a74e9cd905e00e0228433713d735bebd6a9001653b530ac7f3ba84f"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.2.3/go-fuck_4.2.3_darwin_amd64.tar.gz"
      sha256 "60c99a2bac3af03517b456ad6dec4414b9d39b8e028adc5f2030d87db2b4f4d6"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v4.2.3/go-fuck_4.2.3_linux_arm64.tar.gz"
      sha256 "badee9493607aa3ccbd333eb6e6c076be6657cfb192aa9586ae34bdb34e5c053"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v4.2.3/go-fuck_4.2.3_linux_amd64.tar.gz"
      sha256 "ecf4a96aaa46db90dd0c2a55d1b413a46e62289e17bba7b222126edeace8c0c4"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
