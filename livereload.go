package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
)

func main() {
	fmt.Println("Hello, world!")

	startWatching()
}

func startWatching() {
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
			case err := <-watcher.Errors:
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
