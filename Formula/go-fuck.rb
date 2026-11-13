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
  version "2.20.5"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.5/go-fuck_2.20.5_darwin_arm64.tar.gz"
      sha256 "b14665d786ec3292a4cbd3de1a652e065229378ae1b680bcd3b6bc3f64685692"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.5/go-fuck_2.20.5_darwin_amd64.tar.gz"
      sha256 "36ad5d904912092b866e7b284eee64fe19e88aae5b5cc73a03b39825ae906ad3"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.5/go-fuck_2.20.5_linux_arm64.tar.gz"
      sha256 "0f56b75c81eb70e438bcc6a37d6a27789c964c9aec429bdfdcc17eff62af725b"
    else
      url "https://github.com/leafevans/go-fuck/releases/download/v2.20.5/go-fuck_2.20.5_linux_amd64.tar.gz"
      sha256 "906c4e530086c2b4395b2cb6aa3911f1442356bc764f6437323bd73f91e68c4e"
    end
  end

  def install
    bin.install "go-fuck"
  end

  test do
    assert_match "go-fuck #{version}", shell_output("#{bin}/go-fuck --version")
  end
end
