# Candidate processing benchmarks

Run the candidate benchmarks with the toolchain required by `go.mod`:

```sh
go test ./internal/corrector ./internal/ui -run '^$' -bench 'BenchmarkOrganizeCommands|BenchmarkFzfCandidates' -benchmem -count=3
```

The shared command set preserves `CorrectedCommand.Equal` semantics: script and
side effect identify a candidate, while priority and rule name do not. The
corrector preserves its first suggestion, sorts the remaining candidates stably,
and compacts a copy of the input. The fzf selector uses the same set while
keeping display labels separate from executable scripts.

Measurements on Windows amd64 with Go 1.27.4, before and after the shared set:

| Workload                                                                | Bytes/op before | Bytes/op after | Allocations/op before | Allocations/op after |
| ----------------------------------------------------------------------- | --------------: | -------------: | --------------------: | -------------------: |
| Organize 1,000 distinct scripts                                         |          70,992 |         62,800 |                     7 |                    6 |
| Prepare 1,000 fzf candidates                                            |         177,424 |        125,648 |                 1,012 |                   12 |
| Organize 4,000 candidates with shared scripts and distinct side effects |         283,968 |        575,530 |                    19 |                   49 |

The last workload includes duplicates and three distinct side effects per script,
counting the absence of a side effect. Its median time over three runs dropped
from 39.9 ms to 1.06 ms on this host. It trades additional membership storage for
eliminating a scan of all preceding suggestions for each distinct side effect.
The extra index is allocated only when a script has more than one side effect.

Times depend on the host and load. CI guards allocated bytes and object counts
using `baseline.txt`, measured on Linux with the toolchain from `go.mod`.
