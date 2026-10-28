package rules

import (
	"os"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
)

// fixFilePatternSources lists the error formats of the supported tools. The
// order is important: only the first matching format is used.
var fixFilePatternSources = []string{
	// js, node:
	`^    at {file}:{line}:{col}`,
	// cargo:
	`^   {file}:{line}:{col}`,
	// python, thefuck:
	`^  File "{file}", line {line}`,
	// awk:
	`^awk: {file}:{line}:`,
	// git:
	`^fatal: bad config file line {line} in {file}`,
	// llc:
	`^llc: {file}:{line}:{col}:`,
	// lua:
	`^lua: {file}:{line}:`,
	// fish:
	`^{file} \(line {line}\):`,
	// bash, sh, ssh:
	`^{file}: line {line}: `,
	// cargo, clang, gcc, go, pep8, rustc:
	`^{file}:{line}:{col}`,
	// ghc, make, ruby, zsh:
	`^{file}:{line}:`,
	// perl:
	`at {file} line {line}`,
}

// fixFilePattern is a compiled output format.
type fixFilePattern struct {
	re     *regexp.Regexp
	file   int
	line   int
	col    int
	hasCol bool
}

// fixFileCompile compiles one output format.
func fixFileCompile(source string) fixFilePattern {
	pattern := strings.ReplaceAll(source, "{file}", `(?P<file>[^:\n]+)`)
	pattern = strings.ReplaceAll(pattern, "{line}", `(?P<line>[0-9]+)`)
	pattern = strings.ReplaceAll(pattern, "{col}", `(?P<col>[0-9]+)`)

	re := regexp.MustCompile("(?m)" + pattern)
	return fixFilePattern{
		re:     re,
		file:   re.SubexpIndex("file"),
		line:   re.SubexpIndex("line"),
		col:    re.SubexpIndex("col"),
		hasCol: re.SubexpIndex("col") >= 0,
	}
}

// fixFilePatterns holds the compiled output formats.
var fixFilePatterns = func() []fixFilePattern {
	patterns := make([]fixFilePattern, len(fixFilePatternSources))
	for index, source := range fixFilePatternSources {
		patterns[index] = fixFileCompile(source)
	}
	return patterns
}()

// fixFileIsFile reports whether the path is an existing regular file. It is a
// variable so that tests do not depend on the local filesystem.
var fixFileIsFile = func(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// fixFileLocation is a file location extracted from the command output.
type fixFileLocation struct {
	file   string
	line   string
	col    string
	hasCol bool
}

// fixFileSearch returns the first reported location whose file exists.
func fixFileSearch(output string) *fixFileLocation {
	for _, pattern := range fixFilePatterns {
		match := pattern.re.FindStringSubmatch(output)
		if match == nil {
			continue
		}
		file := match[pattern.file]
		if !fixFileIsFile(file) {
			continue
		}
		result := &fixFileLocation{
			file:   file,
			line:   match[pattern.line],
			hasCol: pattern.hasCol,
		}
		if pattern.hasCol {
			result.col = match[pattern.col]
		}
		return result
	}
	return nil
}

// fixFileMatch reports whether the command failed because of an error in a
// file and an editor is configured.
func fixFileMatch(command *types.Command) bool {
	if _, ok := os.LookupEnv("EDITOR"); !ok {
		return false
	}
	return fixFileSearch(command.Output) != nil
}

// fixFileFormat fills the placeholders of an editor command template.
func fixFileFormat(template, editor, file, line, col string) string {
	replacer := strings.NewReplacer(
		"{editor}", editor,
		"{file}", file,
		"{line}", line,
		"{col}", col,
	)
	return replacer.Replace(template)
}

// fixFileNewCommand opens the reported file at the reported line.
func fixFileNewCommand(command *types.Command) ([]string, error) {
	match := fixFileSearch(command.Output)
	if match == nil {
		return nil, nil
	}
	editor := os.Getenv("EDITOR")

	template := conf.Current().String("fixlinecmd", "{editor} {file} +{line}")
	if columnTemplate := conf.Current().String("fixcolcmd", ""); columnTemplate != "" && match.hasCol {
		template = columnTemplate
	}
	editorCall := fixFileFormat(template, editor, match.file, match.line, match.col)
	return []string{shells.Current().And(editorCall, command.Script)}, nil
}

func init() {
	register(types.NewRule("fix_file", fixFileMatch, fixFileNewCommand))
}
