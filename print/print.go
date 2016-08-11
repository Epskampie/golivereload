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

func Line(a ...interface{}) {
	mux.Lock()
	defer mux.Unlock()

	printTime()
	for _, obj := range a {
		fmt.Print(obj, " ")
	}
	fmt.Println("")
}

func Fatal(a ...interface{}) {
	Line(append([]interface{}{red("Error:")}, a...)...)
	os.Exit(1)
}

func printTime() {
	t := time.Now()
	fmt.Print(blue(t.Format("2006/01/02 15:04:05")), " ")
}
