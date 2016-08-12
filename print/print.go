package print

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
)

var mux sync.Mutex

var blue func(a ...interface{}) string = color.New(color.FgBlue).SprintFunc()
var red func(a ...interface{}) string = color.New(color.FgRed).SprintFunc()
var yellow func(a ...interface{}) string = color.New(color.FgYellow).SprintFunc()

var ShowDebug bool = false

func Line(a ...interface{}) {
	mux.Lock()
	defer mux.Unlock()

	printTime()
	for _, obj := range a {
		// fmt.Print(obj, " ")
		fmt.Fprint(color.Output, obj, " ")
	}
	fmt.Println("")
}

func Error(a ...interface{}) {
	Line(append([]interface{}{red("Error:")}, a...)...)
}

func Fatal(a ...interface{}) {
	Error(a...)
	os.Exit(1)
}

func printTime() {
	t := time.Now()
	fmt.Fprint(color.Output, blue(t.Format("2006/01/02 15:04:05")), " ")
}

func Debug(a ...interface{}) {
	if !ShowDebug {
		return
	}
	Line(append([]interface{}{yellow("Debug:")}, a...)...)
}
