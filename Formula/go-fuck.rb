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
  version "2.13.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.13.0/go-fuck_2.13.0_darwin_arm64.tar.gz"
      sha256 "4953492f9bcd223ed32655ccf70c1ebe8b882e24f8d807fb37e1c1dfaa5b49ca"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.13.0/go-fuck_2.13.0_darwin_amd64.tar.gz"
      sha256 "6005c3833277e39fcae0c70d4f4b9cd4098824591876231ff14617b1fc2fbc47"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.13.0/go-fuck_2.13.0_linux_arm64.tar.gz"
      sha256 "89959830a37b36ce5c1c87abbe68df352c98de6f914a2ae69f20940779f8b0d2"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.13.0/go-fuck_2.13.0_linux_amd64.tar.gz"
      sha256 "6952cfee15ae0789239c29f6dc3aafa33cc62777023067b51ad46e82627e0696"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
