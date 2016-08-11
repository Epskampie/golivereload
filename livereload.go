package main

import (
	"encoding/json"
	"flag"
	"livereload/print"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"

	"github.com/fsnotify/fsnotify"
)

var cyan func(a ...interface{}) string = color.New(color.FgCyan).SprintFunc()
var red func(a ...interface{}) string = color.New(color.FgRed).SprintFunc()
var yellow func(a ...interface{}) string = color.New(color.FgYellow).SprintFunc()

func main() {

	flag.Parse()

	// Change rootPath to working dir if not set
	cwd, err := os.Getwd()
	if err != nil {
		print.Fatal(red("Could not get current working dir.", err))
	}
	if params.rootPath == "" {
		params.rootPath = cwd
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

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:

				// print.Line("event:", event)
				// if event.Op&fsnotify.Write == fsnotify.Write {
				// 	print.Line("modified file:", cyan(event.Name))
				// }

				// Watch newly created directories
				if event.Op&fsnotify.Create == fsnotify.Create {
					fileInfo, err := os.Stat(event.Name)
					if err == nil && fileInfo.IsDir() {
						print.Line("got dir, watching:", event.Name)
						err := watcher.Add(event.Name)
						if err != nil {
							log.Fatal(err)
						}
					}
				}

				// Send reload commands
				if event.Op&fsnotify.Write == fsnotify.Write {
					print.Line("reloading:", cyan(event.Name))
					data := reloadRequest{
						Command: "reload",
						Path:    event.Name,
						LiveCSS: strings.HasSuffix(event.Name, ".css"),
					}
					SendJSON <- data
				}
			case err := <-watcher.Errors:
				print.Line("File watcher error:", err)
			}
		}
	}()

	err = watcher.Add(params.rootPath)
	if err == nil {
		print.Line("Watching directory:", cyan(params.rootPath))
	} else {
		print.Line("Got error while writing:")
		log.Fatal(err)
	}
	<-done
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
				print.Line("Got data", string(jsonData))
			}

		} else {
			DelConn <- conn

			return
		}
	}
}

func startServing() {
	print.Line("Start serving")

	http.HandleFunc("/livereload", websocketHandler)
	http.Handle("/", http.FileServer(http.Dir(".")))
	// http.Handle("/", http.FileServer(assetFS()))
	err := http.ListenAndServe(":35729", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
