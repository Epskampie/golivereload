package main

import "flag"

/* Command line parameters */
type ParamsStruct struct {
	includePatterns string
	rootPath        string
}

var params ParamsStruct

func init() {
	flag.StringVar(&params.rootPath, "path", "", "The directory to watch for changes. (default: current directory)")
	flag.StringVar(&params.includePatterns, "include-patterns", "*.html,*.css", "Only reload for files matching these patterns. Not used when -exclude-patterns is defined.")

}
