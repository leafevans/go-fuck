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
- [ ] `gofmt -s -l .` is empty (`make fmt-check`)
- [ ] `make rules-list-check` passes (the rule list in `README.md` is regenerated)
- [ ] No third-party module dependencies were added
- [ ] A new or changed rule has a test, and its fixture is output captured from the real tool
- [ ] The commit messages follow [Conventional Commits](https://www.conventionalcommits.org/)
