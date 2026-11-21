package specific

// ToolAvailability checks the names used by rules for each tool.
type ToolAvailability func(string) bool

// AptAvailable reports whether apt is available.
func (available ToolAvailability) AptAvailable() bool {
	return available("apt-get")
}

// BrewAvailable reports whether Homebrew is available.
func (available ToolAvailability) BrewAvailable() bool {
	return available("brew")
}

// DnfAvailable reports whether dnf is available.
func (available ToolAvailability) DnfAvailable() bool {
	return available("dnf")
}

// YumAvailable reports whether yum is available.
func (available ToolAvailability) YumAvailable() bool {
	return available("yum")
}

// NixAvailable reports whether nix is available.
func (available ToolAvailability) NixAvailable() bool {
	return available("nix")
}

// NpmAvailable reports whether npm is available.
func (available ToolAvailability) NpmAvailable() bool {
	return available("npm")
}

// ChocoAvailable reports whether Chocolatey is available.
func (available ToolAvailability) ChocoAvailable() bool {
	return available("choco") || available("cinst")
}
