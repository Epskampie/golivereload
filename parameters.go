package main

import (
	"flag"
	"fmt"
)

/* Command line parameters */
type ParamsStruct struct {
	includePatterns string
	rootPath        string
	debug           bool
	serve           bool
}

var params ParamsStruct

func init() {
	flag.StringVar(&params.rootPath, "path", "", "The directory to watch for changes.\n    \t(default: current directory)")
	flag.StringVar(&params.includePatterns, "include", "**/*.html,**/*.css", "Only reload for files matching these patterns.")
	flag.BoolVar(&params.debug, "debug", false, "Show debug output.")
	flag.BoolVar(&params.serve, "serve", false, "Serve local webserver that serves files at -path.")

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
