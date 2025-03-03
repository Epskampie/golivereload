package main

import (
	"flag"
	"fmt"
)

/* Command line parameters */
type ParamsStruct struct {
	includePatterns string
	rootPath        string
	cmd             string
	debug           bool
	serve           bool
	version         bool
	delay           int
	port            int
	noLiveCSS       bool
}

var params ParamsStruct

func init() {
	sep := "\n    \t"
	flag.StringVar(&params.rootPath, "path", "", "The directory to watch for changes."+sep+"(default: current directory)")
	flag.StringVar(
		&params.includePatterns,
		"include",
		"**/*.{html,shtml,tmpl,twig,xml,css,js,json}:**/*.{jpeg,jpg,gif,png,ico,cgi}:**/*.{php,py,pl,pm,rb}",
		"Only reload for files matching these patterns."+sep+
			"Use \":\" to separate patterns"+sep+
			"Use \"**\" (double star) to match multiple directories."+sep+
			"Matches are relative to watched path.")
	flag.IntVar(&params.port, "port", 35729, "Port to serve on.")
	flag.BoolVar(&params.debug, "debug", false, "Show debug output.")
	flag.BoolVar(&params.serve, "serve", false, "Start local webserver that serves files at -path.")
	flag.BoolVar(&params.version, "version", false, "Show golivereload version.")
	flag.IntVar(&params.delay, "delay", 0, "Delay this many milliseconds before before sending reload command.")
	flag.StringVar(&params.cmd, "cmd", "", "Command to run after change is detected in files matching pattern. See include-pattern."+sep+"(example: **/*.{scss} ./build.sh)")
	flag.BoolVar(&params.noLiveCSS, "no-live-css", false, "Disable live CSS reloading feature.")

}

func setupFlags(f *flag.FlagSet) {
	f.Usage = func() {
		f.VisitAll(func(f *flag.Flag) {
			s := fmt.Sprintf("  -%s", f.Name) // Two spaces before -; see next two comments.
			name, usage := flag.UnquoteUsage(f)
			if len(name) > 0 {
				s += " " + name
			}
			// Four spaces before the tab triggers good alignment
			// for both 4- and 8-space tab stops.
			s += "\n    \t"

			s += usage
			if !isZeroValue(f.DefValue) {
				s += fmt.Sprintf("\n    \t(default %v)", f.DefValue)
			}
			fmt.Println(s)
		})
	}
}

// isZeroValue guesses whether the string represents the zero
// value for a flag. It is not accurate but in practice works OK.
func isZeroValue(value string) bool {
	switch value {
	case "false":
		return true
	case "":
		return true
	case "0":
		return true
	}
	return false
}
