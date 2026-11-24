package rules

import (
	"path/filepath"
	"strings"
)

// commandInvocation reports whether script runs name as a command: either as the
// program itself or as the "-name" subcommand of hdfs. A word that only appears
// as an argument of another command, such as mkdir in "grep -rn mkdir src", is
// not an invocation.
func commandInvocation(script, name string) bool {
	fields := strings.Fields(script)
	if len(fields) == 0 {
		return false
	}
	if filepath.Base(fields[0]) == name {
		return true
	}
	subcommand := "-" + name
	for index := 1; index < len(fields); index++ {
		if fields[index] != subcommand {
			continue
		}
		switch filepath.Base(fields[index-1]) {
		case "hdfs", "dfs":
			return true
		}
	}
	return false
}
