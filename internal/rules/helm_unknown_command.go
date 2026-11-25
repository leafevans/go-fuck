package rules

func init() {
	registerScopedCobraRule("helm_unknown_command", "helm", []string{"helm", "helm.exe"}, func(name string) commandOption {
		switch name {
		case "--namespace", "-n", "--kube-context", "--kubeconfig":
			return valueOption
		case "--debug":
			return flagOption
		default:
			return unknownOption
		}
	})
}
