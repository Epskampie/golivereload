package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"golivereload/print"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar"
	"github.com/fatih/color"
	"github.com/gorilla/websocket"
	"github.com/rjeczalik/notify"
)

var cyan func(a ...interface{}) string = color.New(color.FgCyan).SprintFunc()
var red func(a ...interface{}) string = color.New(color.FgRed).SprintFunc()
var yellow func(a ...interface{}) string = color.New(color.FgYellow).SprintFunc()

func main() {

	setupFlags(flag.CommandLine)
	flag.Parse()

	if params.version {
		fmt.Println("0.2.0")
		os.Exit(0)
	}

	print.ShowDebug = params.debug

	// Change rootPath to working dir if not set
	cwd, err := os.Getwd()
	if err != nil {
		print.Fatal(red("Could not get current working dir.", err))
	}
	if params.rootPath == "" {
		params.rootPath = cwd
	}

	if !strings.HasSuffix(params.rootPath, string(os.PathSeparator)) {
		params.rootPath += string(os.PathSeparator)
	}

	// Check rootPath
	fileInfo, err := os.Stat(params.rootPath)
	if err != nil {
		print.Fatal(red(err))
	}
	if !fileInfo.IsDir() {
		print.Fatal(cyan(params.rootPath), red("is not a directory"))
	}

	go watchFilesystem()
	go StartWebsocketPool()
	startServing()
}

/* ======= Filesytem watching ======= */

func watchFilesystem() {
	prevTime := time.Now().UnixNano()
	prevName := ""
	includePatterns := strings.Split(params.includePatterns, ":")

	notifyChannel := make(chan notify.EventInfo, 1)

	if err := notify.Watch(params.rootPath+"...", notifyChannel, notify.All); err != nil {
		log.Fatal(err)
	} else {
		print.Line("Watching directory:", cyan(params.rootPath))
		print.Line("Using include patterns:", params.includePatterns)
	}
	defer notify.Stop(notifyChannel)

WATCHLOOP:
	for {
		switch event := <-notifyChannel; event.Event() {
		case notify.Write:

			// De-duplicate event
			now := time.Now().UnixNano()
			duplicate := event.Path() == prevName && (now-prevTime) < int64(100*time.Millisecond)
			prevTime = now
			prevName = event.Path()
			if duplicate {
				print.Debug("De-duplicated event")
				continue WATCHLOOP
			}

			trimmedPath := strings.TrimPrefix(event.Path(), params.rootPath)

			go processEvent(trimmedPath, event, includePatterns)
		default:
			print.Debug("Got event", event)
		}
	}
}

func tryMatch(patterns []string, subject string) (matched bool, pattern string) {
	for _, pattern := range patterns {
		print.Debug("Pattern", pattern, "Path", subject)
		match, err := doublestar.Match(pattern, subject)
		if err != nil {
			print.Error("Invalid pattern:", err)
		}
		if match {
			return true, pattern
		}
	}
	return false, ""
}

func processEvent(trimmedPath string, event notify.EventInfo, includePatterns []string) {
	// Run cmd
	if params.cmd != "" {
		splitCmd := strings.Split(params.cmd, " ")

		matched, _ := tryMatch([]string{splitCmd[0]}, trimmedPath)
		if matched {
			print.Line("Running cmd: ", splitCmd[1:])
			time.Sleep(50 * time.Millisecond)

			cmd := exec.Command(splitCmd[1], splitCmd[2:]...)
			err := cmd.Run()
			if err != nil {
				print.Error("Error running cmd: ", err)
			}
		}
	}

	// Run cmd
	if len(includePatterns) > 0 {
		matched, pattern := tryMatch(includePatterns, trimmedPath)
		if matched {
			print.Debug("Match found", pattern, trimmedPath)
		} else {
			print.Line(yellow("Ignoring:"), cyan(trimmedPath))
			return
		}
	}

	print.Line("Reloading:", cyan(trimmedPath))
	if params.delay > 0 {
		print.Line("Delaying", params.delay, "ms first")
		time.Sleep(time.Duration(params.delay) * time.Millisecond)
		print.Line("sending")
	}
	data := reloadRequest{
		Command: "reload",
		Path:    event.Path(),
		LiveCSS: strings.HasSuffix(event.Path(), ".css"),
	}
	SendJSON <- data
}

/* ======= Websockets ======= */

type reloadRequest struct {
	Command string `json:"command"`
	Path    string `json:"path"`
	LiveCSS bool   `json:"liveCSS"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     CheckOrigin,
}

func CheckOrigin(r *http.Request) bool {
	return true
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		print.Line(err)
		return
	}

	AddConn <- conn

	defer conn.Close()

	for {
		var data map[string]interface{}
		err := conn.ReadJSON(&data)

		if err == nil {

			/* Hello command*/
			if data["command"] == "hello" {
				SendString <- "{\"command\":\"hello\",\"protocols\":[\"http://livereload.com/protocols/official-7\",\"http://livereload.com/protocols/official-8\",\"http://livereload.com/protocols/official-9\",\"http://livereload.com/protocols/2.x-origin-version-negotiation\",\"http://livereload.com/protocols/2.x-remote-control\"],\"serverName\":\"LiveReload 2\"}"
			} else {
				jsonData, _ := json.Marshal(&data)
				print.Debug("Got data", string(jsonData))
			}

		} else {
			DelConn <- conn

			return
		}
	}
}

func startServing() {
	port := strconv.Itoa(params.port)

	changeHeaderThenServe := func(h http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Set some header.
			w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
			// Serve with the actual handler.
			h.ServeHTTP(w, r)
		}
	}

	http.HandleFunc("/livereload", websocketHandler)
	http.Handle("/livereload.js", http.FileServer(assetFS()))
	if params.serve {
		print.Line("Serving files from:", cyan(params.rootPath), "on:", cyan("http://localhost:"+port))
		http.Handle("/", changeHeaderThenServe(http.FileServer(http.Dir(params.rootPath))))
	} else {
		print.Line("Listening on port:", port)
	}
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		if _, ok := err.(*net.OpError); ok {
			print.Fatal("Port", port, "already in use")
		} else {
			print.Fatal("ListenAndServe: " + err.Error())
		}
	}
}
