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

## Rule matching on long captured output

`internal/rules` benchmarks the rule scan for a short capture
(`BenchmarkRuleMatchingAllRules`) and for a 4,000 line capture
(`BenchmarkRuleMatchingLongOutput` with `build` and `git` sub-benchmarks).
The long benchmark exists because several rules inspect every captured byte: a
rule that scans a 200 KB capture without a cheap literal check shows up there,
while it is invisible with a normal failure message.
`BenchmarkMissingSpaceExecutable` covers the executable prefix scan of the
missing space rule on a synthetic 5,000 name list.

Run them with:

```sh
go test ./internal/rules -run '^$' -bench 'BenchmarkRuleMatchingLongOutput|BenchmarkMissingSpaceExecutable' -benchmem
```

Measurements on Windows amd64 with Go 1.27.4, medians of three runs, before and
after the two rule scan pull requests:

| Workload                                | v4.2.2                                   | after #20                               | after #23 (v4.2.4)                      |
| --------------------------------------- | ---------------------------------------: | --------------------------------------: | --------------------------------------: |
| `BenchmarkRuleMatchingLongOutput/build` | 19.82 ms/op, 430,940 B/op, 20 allocs/op | 1.82 ms/op, 217,300 B/op, 17 allocs/op | 1.64 ms/op, 217,100 B/op, 15 allocs/op |
| `BenchmarkRuleMatchingLongOutput/git`   | 37.12 ms/op, 710,622 B/op, 23 allocs/op | 5.06 ms/op, 430,800 B/op, 18 allocs/op | 2.94 ms/op, 430,600 B/op, 16 allocs/op |

The remaining allocated bytes are the rules that still lower case the capture
(`sudo` and `git_hook_bypass`). Times depend on the host; `baseline.txt`
guards allocated bytes and allocations, not time.
