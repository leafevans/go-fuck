## Changelog

- Preserve valid history commands such as `fd` and `fzf` when the correction alias is `f`. History filtering now compares complete command names and accepts leading whitespace and tab-separated arguments.
- Exclude failed history commands followed by correction calls with arguments, such as `f --yes`.
- Share candidate identity tracking between the corrector and fzf, reduce allocations for ordinary suggestions, and avoid quadratic scans when distinct side effects share the same script. Candidate order, side-effect markers, and executable scripts are preserved.
- Make the tcsh integration check load the generated alias in tcsh and verify that it was defined, preventing an empty alias from passing validation.
