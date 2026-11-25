package rules

func init() {
	registerScopedCobraRule("kubectl_unknown_command", "kubectl", []string{"kubectl", "kubectl.exe"}, func(name string) commandOption {
		switch name {
		case "--context", "--namespace", "-n", "--kubeconfig", "--cluster", "--user", "--server", "-s":
			return valueOption
		case "--warnings-as-errors", "--insecure-skip-tls-verify":
			return flagOption
		default:
			return unknownOption
		}
	})
}
