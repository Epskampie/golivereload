package main

import (
	"encoding/json"
	"flag"
	"livereload/print"
	"log"
	"net/http"
	"os"
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
	includePatterns := strings.Split(params.includePatterns, ",")

	notifyChannel := make(chan notify.EventInfo, 1)

	if err := notify.Watch(params.rootPath+"...", notifyChannel, notify.All); err != nil {
		log.Fatal(err)
	} else {
		print.Line("Watching directory:", cyan(params.rootPath))
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

			// Send reload commands

			displayName := strings.TrimPrefix(event.Path(), params.rootPath)

			if len(includePatterns) > 0 {
				matched := false
				for _, pattern := range includePatterns {
					print.Debug("Pattern", pattern)
					match, err := doublestar.Match(pattern, event.Path())
					if err != nil {
						print.Error("Invalid pattern:", err)
					}
					if match {
						print.Debug("Match found", pattern, event.Path())
						matched = true
						break
					}
				}
				if !matched {
					print.Line(yellow("Ignoring:"), cyan(displayName))
					continue WATCHLOOP
				}
			}

			print.Line("Reloading:", cyan(displayName))
			data := reloadRequest{
				Command: "reload",
				Path:    event.Path(),
				LiveCSS: strings.HasSuffix(event.Path(), ".css"),
			}
			SendJSON <- data
		default:
			print.Debug("Got event", event)
		}
	}
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
	port := "35729"
	print.Line("Listening on port:", port)

	http.HandleFunc("/livereload", websocketHandler)
	http.Handle("/livereload.js", http.FileServer(assetFS()))
	// http.Handle("/", http.FileServer(http.Dir(".")))
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
