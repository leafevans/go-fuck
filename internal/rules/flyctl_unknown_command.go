package rules

func init() {
	registerScopedCobraRule("flyctl_unknown_command", "fly", []string{"flyctl", "fly", "flyctl.exe", "fly.exe"}, func(name string) commandOption {
		switch name {
		case "--org", "--app", "-a", "--config", "-c", "--access-token", "-t":
			return valueOption
		case "--debug", "--verbose":
			return flagOption
		default:
			return unknownOption
		}
	})
}
