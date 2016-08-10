package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"

	"github.com/fsnotify/fsnotify"
)

var connections []*websocket.Conn

type reloadRequest struct {
	Command string `json:"command"`
	Path    string `json:"path"`
	LiveCSS bool   `json:"liveCSS"`
}

func main() {
	go startWatching()
	startServing()
}

func startWatching() {
	fmt.Println("Start watching")

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

				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
				}

				// Watch newly created directories
				if event.Op&fsnotify.Create == fsnotify.Create {
					fileInfo, err := os.Stat(event.Name)
					if err == nil && fileInfo.IsDir() {
						log.Println("got dir, watching:", event.Name)
						err := watcher.Add(event.Name)
						if err != nil {
							log.Fatal(err)
						}
					}
				}

				// Send reload commands
				if event.Op&fsnotify.Write == fsnotify.Write {
					fmt.Println("Sending write", event.Name)
					for _, conn := range connections {
						data := reloadRequest{
							Command: "reload",
							Path:    event.Name,
							// liveCSS: strings.HasSuffix(event.Name, ".css"),
							LiveCSS: false,
						}
						dataString, err := json.Marshal(data)

						fmt.Println(string(dataString), err)
						err = conn.WriteJSON(data)

						if err != nil {
							fmt.Println(err)
						} else {
							fmt.Println("Sent write to")
						}
					}
				}
			case err := <-watcher.Errors:
				// os.Stat(event)
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add("/home/simon/tmp")
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     CheckOrigin,
}

func CheckOrigin(r *http.Request) bool {
	return true
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	connections = append(connections, conn)

	for {
		var data map[string]interface{}
		err := conn.ReadJSON(&data)

		if err == nil {
			jsonData, err := json.Marshal(&data)
			fmt.Println("Got data", string(jsonData), err)

			// Helo
			if data["command"] == "hello" {
				fmt.Println("Got hello command")
				writeString(conn, "{\"command\":\"hello\",\"protocols\":[\"http://livereload.com/protocols/official-7\",\"http://livereload.com/protocols/official-8\",\"http://livereload.com/protocols/official-9\",\"http://livereload.com/protocols/2.x-origin-version-negotiation\",\"http://livereload.com/protocols/2.x-remote-control\"],\"serverName\":\"LiveReload 2\"}")
			}

		} else {
			fmt.Println(err)

			return
		}
	}
}

func writeString(conn *websocket.Conn, msg string) {
	err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		fmt.Println(err)
	}
}

func startServing() {
	fmt.Println("Start serving")

	http.HandleFunc("/livereload", echoHandler)
	http.Handle("/", http.FileServer(http.Dir(".")))
	err := http.ListenAndServe(":35729", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
