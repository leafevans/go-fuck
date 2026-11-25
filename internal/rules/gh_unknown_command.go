package rules

func init() {
	registerScopedCobraRule("gh_unknown_command", "gh", []string{"gh", "gh.exe"}, func(string) commandOption {
		return unknownOption
	})
}
