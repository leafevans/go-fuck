## Summary

<!-- What does this change and why? Reference the issue it fixes, if any. -->

## Type of change

- [ ] Bug fix
- [ ] New rule
- [ ] Rule fix or refresh
- [ ] Documentation
- [ ] Refactor or tooling

## Checklist

- [ ] `go test ./... -count=1` passes
- [ ] `go vet ./...` passes
- [ ] `task vuln` and `task gosec` pass
- [ ] `gofmt -s -l .` is empty (`task fmt-check`)
- [ ] `task rules-list-check` passes (the rule list in `README.md` is regenerated)
- [ ] No third-party module dependencies were added
- [ ] A new or changed rule has a test, and its fixture is output captured from the real tool
- [ ] User-facing changes update `README.md` and `README.zh-CN.md` when relevant
- [ ] The commit messages follow [Conventional Commits](https://www.conventionalcommits.org/)
