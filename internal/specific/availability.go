package specific

import (
	"github.com/leafevans/go-fuck/internal/utils"
)

// AptAvailable reports whether apt is available.
func AptAvailable() bool {
	return utils.Which("apt-get") != ""
}

// BrewAvailable reports whether Homebrew is available.
func BrewAvailable() bool {
	return utils.Which("brew") != ""
}

// DnfAvailable reports whether dnf is available.
func DnfAvailable() bool {
	return utils.Which("dnf") != ""
}

// YumAvailable reports whether yum is available.
func YumAvailable() bool {
	return utils.Which("yum") != ""
}

// NixAvailable reports whether nix is available.
func NixAvailable() bool {
	return utils.Which("nix") != ""
}

// NpmAvailable reports whether npm is available.
func NpmAvailable() bool {
	return utils.Which("npm") != ""
}

// ChocoAvailable reports whether Chocolatey is available.
func ChocoAvailable() bool {
	return utils.Which("choco") != "" || utils.Which("cinst") != ""
}
