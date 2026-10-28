package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitCloneMissingErrors lists the outputs that report an unknown command.
var gitCloneMissingErrors = []string{
	"No such file or directory",
	"not found",
	"is not recognised as",
}

// gitCloneMissingASCIIAlpha reports whether a byte is an ASCII letter.
func gitCloneMissingASCIIAlpha(char byte) bool {
	return char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z'
}

// gitCloneMissingSchemeChar reports whether a byte may appear in an URL
// scheme.
func gitCloneMissingSchemeChar(char byte) bool {
	return gitCloneMissingASCIIAlpha(char) || char >= '0' && char <= '9' ||
		char == '+' || char == '-' || char == '.'
}

// gitCloneMissingURLScheme parses the scheme and network location of a URL the
// same way as Python's urllib.parse.urlparse with a default "ssh" scheme.
func gitCloneMissingURLScheme(script string) (scheme, netloc string) {
	scheme, rest := "ssh", script
	if index := strings.Index(script, ":"); index > 0 && gitCloneMissingASCIIAlpha(script[0]) {
		candidate := script[:index]
		valid := true
		for offset := 0; offset < len(candidate); offset++ {
			if !gitCloneMissingSchemeChar(candidate[offset]) {
				valid = false
				break
			}
		}
		if valid {
			scheme = strings.ToLower(candidate)
			rest = script[index+1:]
		}
	}

	if strings.HasPrefix(rest, "//") {
		rest = rest[2:]
		if end := strings.IndexAny(rest, "/?#"); end >= 0 {
			netloc = rest[:end]
		} else {
			netloc = rest
		}
	}
	return scheme, netloc
}

// gitCloneMissingMatch reports whether a bare URL was run as a command.
func gitCloneMissingMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	if len(parts) != 1 {
		return false
	}
	if utils.Which(parts[0]) != "" {
		return false
	}

	knownError := false
	for _, errorOutput := range gitCloneMissingErrors {
		if strings.Contains(command.Output, errorOutput) {
			knownError = true
			break
		}
	}
	if !knownError {
		return false
	}

	scheme, netloc := gitCloneMissingURLScheme(command.Script)
	if netloc == "" && scheme != "ssh" {
		return false
	}
	if scheme == "ssh" &&
		(!strings.Contains(command.Script, "@") || !strings.Contains(command.Script, ":")) {
		return false
	}
	return scheme == "http" || scheme == "https" || scheme == "ssh"
}

// gitCloneMissingNewCommand clones the pasted URL.
func gitCloneMissingNewCommand(command *types.Command) ([]string, error) {
	return []string{"git clone " + command.Script}, nil
}

func init() {
	register(types.NewRule("git_clone_missing",
		gitCloneMissingMatch,
		gitCloneMissingNewCommand))
}
