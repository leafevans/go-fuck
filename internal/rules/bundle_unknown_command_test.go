package rules

import (
	"slices"
	"sort"
	"strconv"
	"testing"
)

// bundleUnknownCommandHelpOutput is the output of "bundle help" of Bundler
// 4.0.20 when no system man page viewer is installed. Bundler then prints the
// raw ronn source of the bundle(1) manual page, so every command reference is
// backticked. Copied from lib/bundler/man/bundle.1.ronn.
const bundleUnknownCommandHelpOutput = "bundle(1) -- Ruby Dependency Management\n" +
	"=======================================\n" +
	"\n" +
	"## SYNOPSIS\n" +
	"\n" +
	"`bundle` COMMAND [--no-color] [--verbose] [ARGS]\n" +
	"\n" +
	"## DESCRIPTION\n" +
	"\n" +
	"Bundler manages an `application's dependencies` through its entire life\n" +
	"across many machines systematically and repeatably.\n" +
	"\n" +
	"See [the RubyGems guides](https://guides.rubygems.org/) for information on getting\n" +
	"started, and Gemfile(5) for more information on the `Gemfile` format.\n" +
	"\n" +
	"## OPTIONS\n" +
	"\n" +
	"* `--no-color`:\n" +
	"  Print all output without color\n" +
	"\n" +
	"* `--retry`, `-r`:\n" +
	"  Specify the number of times you wish to attempt network commands\n" +
	"\n" +
	"* `--verbose`, `-V`:\n" +
	"  Print out additional logging information\n" +
	"\n" +
	"## BUNDLE COMMANDS\n" +
	"\n" +
	"We divide `bundle` subcommands into primary commands and utilities:\n" +
	"\n" +
	"## PRIMARY COMMANDS\n" +
	"\n" +
	"* [`bundle install(1)`](bundle-install.1.html):\n" +
	"  Install the gems specified by the `Gemfile` or `Gemfile.lock`\n" +
	"\n" +
	"* [`bundle update(1)`](bundle-update.1.html):\n" +
	"  Update dependencies to their latest versions\n" +
	"\n" +
	"* [`bundle cache(1)`](bundle-cache.1.html):\n" +
	"  Package the .gem files required by your application into the\n" +
	"  `vendor/cache` directory (aliases: `bundle package`, `bundle pack`)\n" +
	"\n" +
	"* [`bundle exec(1)`](bundle-exec.1.html):\n" +
	"  Execute a script in the current bundle\n" +
	"\n" +
	"* [`bundle config(1)`](bundle-config.1.html):\n" +
	"  Specify and read configuration options for Bundler\n" +
	"\n" +
	"* [`bundle help(1)`](bundle-help.1.html):\n" +
	"  Display detailed help for each subcommand\n" +
	"\n" +
	"## UTILITIES\n" +
	"\n" +
	"* [`bundle add(1)`](bundle-add.1.html):\n" +
	"  Add the named gem to the Gemfile and run `bundle install`\n" +
	"\n" +
	"* [`bundle binstubs(1)`](bundle-binstubs.1.html):\n" +
	"  Generate binstubs for executables in a gem\n" +
	"\n" +
	"* [`bundle check(1)`](bundle-check.1.html):\n" +
	"  Determine whether the requirements for your application are installed\n" +
	"  and available to Bundler\n" +
	"\n" +
	"* [`bundle show(1)`](bundle-show.1.html):\n" +
	"  Show the source location of a particular gem in the bundle\n" +
	"\n" +
	"* [`bundle info(1)`](bundle-info.1.html):\n" +
	"  Show information for the given gem in your bundle\n" +
	"\n" +
	"* [`bundle list(1)`](bundle-list.1.html):\n" +
	"  List all the gems in the bundle\n" +
	"\n" +
	"* [`bundle licenses(1)`](bundle-licenses.1.html):\n" +
	"  Print the license of all gems in the bundle\n" +
	"\n" +
	"* [`bundle fund(1)`](bundle-fund.1.html):\n" +
	"  Lists information about gems seeking funding assistance\n" +
	"\n" +
	"* [`bundle outdated(1)`](bundle-outdated.1.html):\n" +
	"  Show all of the outdated gems in the current bundle\n" +
	"\n" +
	"* `bundle console(1)` (deprecated):\n" +
	"  Start an IRB session in the current bundle\n" +
	"\n" +
	"* [`bundle open(1)`](bundle-open.1.html):\n" +
	"  Open an installed gem in the editor\n" +
	"\n" +
	"* [`bundle lock(1)`](bundle-lock.1.html):\n" +
	"  Generate a lockfile for your dependencies\n" +
	"\n" +
	"* [`bundle init(1)`](bundle-init.1.html):\n" +
	"  Generate a simple `Gemfile`, placed in the current directory\n" +
	"\n" +
	"* [`bundle gem(1)`](bundle-gem.1.html):\n" +
	"  Create a simple gem, suitable for development with Bundler\n" +
	"\n" +
	"* [`bundle platform(1)`](bundle-platform.1.html):\n" +
	"  Display platform compatibility information\n" +
	"\n" +
	"* [`bundle clean(1)`](bundle-clean.1.html):\n" +
	"  Clean up unused gems in your Bundler directory\n" +
	"\n" +
	"* [`bundle pristine(1)`](bundle-pristine.1.html):\n" +
	"  Restores installed gems to their pristine condition\n" +
	"\n" +
	"* [`bundle doctor(1)`](bundle-doctor.1.html):\n" +
	"  Display warnings about common problems\n" +
	"\n" +
	"* [`bundle env(1)`](bundle-env.1.html):\n" +
	"  Print information about the environment Bundler is running under\n" +
	"\n" +
	"* [`bundle issue(1)`](bundle-issue.1.html):\n" +
	"  Get help reporting Bundler issues\n" +
	"\n" +
	"* [`bundle remove(1)`](bundle-remove.1.html):\n" +
	"  Removes gems from the Gemfile\n" +
	"\n" +
	"* [`bundle plugin(1)`](bundle-plugin.1.html):\n" +
	"  Manage Bundler plugins\n" +
	"\n" +
	"* [`bundle version(1)`](bundle-version.1.html):\n" +
	"  Prints Bundler version information\n" +
	"\n" +
	"## PLUGINS\n" +
	"\n" +
	"When running a command that isn't listed in PRIMARY COMMANDS or UTILITIES,\n" +
	"Bundler will try to find an executable on your path named `bundler-<command>`\n" +
	"and execute it, passing down any extra arguments to it.\n" +
	""

// bundleUnknownCommandWantCommands is the command list of the help output
// above, including the aliases of the primary commands.
var bundleUnknownCommandWantCommands = []string{
	"install", "update", "cache", "package", "pack", "exec", "config",
	"help", "add", "binstubs", "check", "show", "info", "list", "licenses",
	"fund", "outdated", "console", "open", "lock", "init", "gem", "platform",
	"clean", "pristine", "doctor", "env", "issue", "remove", "plugin",
	"version",
}

// bundleUnknownCommandOutput returns the output Bundler prints for an unknown
// command.
func bundleUnknownCommandOutput(name string) string {
	return "Could not find command " + strconv.Quote(name) + "."
}

func TestRuleBundleUnknownCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("bundle_unknown_command")
	if rule == nil {
		t.Fatal("bundle_unknown_command rule is not registered")
	}

	t.Run("parse_commands", func(t *testing.T) {
		got := bundleUnknownCommandParseCommands(bundleUnknownCommandHelpOutput)
		if !slices.Equal(got, bundleUnknownCommandWantCommands) {
			t.Errorf("parse commands = %q, want %q", got, bundleUnknownCommandWantCommands)
		}
	})

	t.Run("fallback_commands", func(t *testing.T) {
		fallback := append([]string{}, bundleUnknownCommandFallbackCommands...)
		parsed := append([]string{}, bundleUnknownCommandWantCommands...)
		sort.Strings(fallback)
		sort.Strings(parsed)
		if !slices.Equal(fallback, parsed) {
			t.Errorf("fallback commands = %q, want the help commands %q", fallback, parsed)
		}
	})

	originalHelp := bundleUnknownCommandHelp
	t.Cleanup(func() { bundleUnknownCommandHelp = originalHelp })
	bundleUnknownCommandHelp = func() []string {
		return bundleUnknownCommandParseCommands(bundleUnknownCommandHelpOutput)
	}

	assertMatch(t, rule, "bundle zzz", bundleUnknownCommandOutput("zzz"))
	assertMatch(t, rule, "bundle instal", bundleUnknownCommandOutput("instal"))
	assertNotMatch(t, rule, "bundle zzz", "")
	assertNotMatch(t, rule, "bundle install", "Bundle complete! 2 Gemfile dependencies, 5 gems now installed.")
	assertNotMatch(t, rule, "gem zzz", bundleUnknownCommandOutput("zzz"))

	commandCases := []struct {
		name   string
		script string
		output string
		want   []string
	}{
		{"unknown", "bundle zzz", bundleUnknownCommandOutput("zzz"), []string{}},
		{"misspelled", "bundle isntall", bundleUnknownCommandOutput("isntall"), []string{"bundle install"}},
		{"arguments", "bundle isntall --local", bundleUnknownCommandOutput("isntall"), []string{"bundle install --local"}},
		{"several", "bundle updat", bundleUnknownCommandOutput("updat"), []string{"bundle update", "bundle outdated"}},
	}
	for _, testCase := range commandCases {
		t.Run("new_command/"+testCase.name, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script, testCase.output, testCase.want)
		})
	}
}
