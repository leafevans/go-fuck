package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/system"
)

// whichCacheEntry is a memoized executable lookup. The entry is only valid
// while PATH is unchanged, so tests that modify PATH keep working.
type whichCacheEntry struct {
	path    string
	pathEnv string
}

var whichCache sync.Map

// Which returns the absolute path of an executable, or an empty string when it
// is not available.
//
// The lookup is memoized because rule matching asks for the same programs many
// times and every miss costs a number of file system probes on Windows.
func Which(program string) string {
	pathEnv := os.Getenv("PATH")
	if cached, ok := whichCache.Load(program); ok {
		entry := cached.(whichCacheEntry)
		if entry.pathEnv == pathEnv {
			return entry.path
		}
	}

	resolved := system.Which(program)
	whichCache.Store(program, whichCacheEntry{path: resolved, pathEnv: pathEnv})
	return resolved
}

// ResetCaches clears the process wide caches. It is intended for tests.
func ResetCaches() {
	executablesOnce = sync.Once{}
	executables = nil
	executableSet = nil
	whichCache = sync.Map{}
	pathScanMu.Lock()
	pathScanCache = nil
	pathScanPath = ""
	pathScanMu.Unlock()
}

var (
	executablesOnce sync.Once
	executables     []string
	executableSet   map[string]bool
)

// GetAllExecutables returns the names of all executables that are available in
// PATH, together with the shell aliases.
func GetAllExecutables() []string {
	executablesOnce.Do(func() {
		executables = loadAllExecutables()
		executableSet = make(map[string]bool, len(executables))
		for _, name := range executables {
			executableSet[name] = true
		}
	})
	return executables
}

// IsExecutable reports whether name is available in PATH or as a shell alias.
func IsExecutable(name string) bool {
	GetAllExecutables()
	return executableSet[name]
}

// executableEntryPoints are the go-fuck binaries that must not be suggested as
// corrected commands.
var executableEntryPoints = map[string]bool{
	"go-fuck": true,
	"thefuck": true,
	"fuck":    true,
}

// pathScan is the cached result of scanning the PATH directories.
type pathScan struct {
	// names are the executable file names in PATH order.
	names []string
	// available maps normalized program names to true.
	available map[string]bool
}

var (
	pathScanMu    sync.Mutex
	pathScanPath  string
	pathScanCache *pathScan
)

// scanPath returns the cached PATH scan.
//
// The cache is keyed by PATH so that tests and shell history helpers see tools
// that were added to PATH later in the process.
func scanPath() *pathScan {
	path := os.Getenv("PATH")
	key := path + "\x00" + strings.Join(conf.Current().ExcludedSearchPathPrefixes, "\x00")

	pathScanMu.Lock()
	defer pathScanMu.Unlock()
	if pathScanCache != nil && pathScanPath == key {
		return pathScanCache
	}

	names := readPathExecutableNames(path)
	available := make(map[string]bool, len(names))
	for _, name := range names {
		if normalized := availableFileName(name); normalized != "" {
			available[normalized] = true
		}
	}
	pathScanPath = key
	pathScanCache = &pathScan{names: names, available: available}
	return pathScanCache
}

// IsAvailable reports whether an executable called program exists in PATH.
//
// Unlike Which it does not resolve the absolute path. The PATH directories are
// scanned once and cached, so checking the availability of many programs costs
// a single directory scan instead of one executable lookup per program. This is
// what makes the enabled_by_default flags of the rules cheap to evaluate.
func IsAvailable(program string) bool {
	return scanPath().available[availableName(program)]
}

// availableFileName normalizes the name of a scanned PATH file for
// availability checks. It returns an empty string when the file cannot be
// executed on the platform.
func availableFileName(name string) string {
	lower := strings.ToLower(name)
	if runtime.GOOS != "windows" {
		// LookPath would also check the executable bit; the directory scan does
		// not, which can only make a rule enabled when the tool is not runnable.
		return lower
	}
	for _, extension := range windowsExecutableExtensions {
		if strings.HasSuffix(lower, extension) {
			return strings.TrimSuffix(lower, extension)
		}
	}
	return ""
}

// availableName normalizes a program name for availability lookups.
func availableName(program string) string {
	lower := strings.ToLower(program)
	if runtime.GOOS != "windows" {
		return lower
	}
	for _, extension := range windowsExecutableExtensions {
		if strings.HasSuffix(lower, extension) {
			return strings.TrimSuffix(lower, extension)
		}
	}
	return lower
}

// windowsExecutableExtensions are the executable suffixes that are considered
// available on Windows.
var windowsExecutableExtensions = [...]string{".exe", ".cmd", ".bat", ".com"}

// readPathExecutableNames scans the PATH directories in parallel and returns
// the file names of all executable candidates, in PATH order.
//
// The scan is parallel because a typical Windows PATH contains many
// directories and the scan is latency bound.
func readPathExecutableNames(path string) []string {
	alias := GetAlias()
	aliasLower := strings.ToLower(alias)

	paths := filepath.SplitList(path)
	dirs := make([]string, 0, len(paths))
	for _, dir := range paths {
		if dir != "" && includePathInSearch(dir) {
			dirs = append(dirs, dir)
		}
	}

	perDir := make([][]string, len(dirs))
	if len(dirs) > 0 {
		workers := min(runtime.NumCPU(), len(dirs))
		jobs := make(chan int)
		var waitGroup sync.WaitGroup
		waitGroup.Add(workers)
		for range workers {
			go func() {
				defer waitGroup.Done()
				for index := range jobs {
					perDir[index] = readExecutableDir(dirs[index], aliasLower)
				}
			}()
		}
		for index := range dirs {
			jobs <- index
		}
		close(jobs)
		waitGroup.Wait()
	}

	names := []string{}
	for _, dirNames := range perDir {
		names = append(names, dirNames...)
	}
	return names
}

// loadAllExecutables returns the PATH executables and the shell aliases.
func loadAllExecutables() []string {
	found := append([]string{}, scanPath().names...)
	alias := GetAlias()
	for name := range shells.Current().GetAliases() {
		if name != alias {
			found = append(found, name)
		}
	}
	return found
}

// readExecutableDir returns the names of the executable files of a PATH
// directory. The result order matches os.ReadDir, so the flattened list is
// deterministic and follows PATH order.
func readExecutableDir(dir, aliasLower string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		base := strings.ToLower(trimExecutableSuffix(name))
		if executableEntryPoints[base] || base == aliasLower {
			continue
		}
		names = append(names, name)
	}
	return names
}

// trimExecutableSuffix removes a Windows executable suffix.
func trimExecutableSuffix(name string) string {
	if len(name) > 4 && strings.EqualFold(name[len(name)-4:], ".exe") {
		return name[:len(name)-4]
	}
	return name
}

// includePathInSearch reports whether a PATH entry should be searched.
func includePathInSearch(path string) bool {
	for _, prefix := range conf.Current().ExcludedSearchPathPrefixes {
		if prefix != "" && strings.HasPrefix(path, prefix) {
			return false
		}
	}
	return true
}
