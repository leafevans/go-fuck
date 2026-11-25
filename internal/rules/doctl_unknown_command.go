package rules

func init() {
	registerScopedCobraRule("doctl_unknown_command", "doctl", []string{"doctl", "doctl.exe"}, func(name string) commandOption {
		switch name {
		case "--context", "--access-token", "-t", "--config", "-c", "--output", "-o":
			return valueOption
		case "--trace", "--verbose", "-v":
			return flagOption
		default:
			return unknownOption
		}
	})
}
