package main

import (
	"encoding/json"
	"golivereload/print"
	"time"

	"github.com/gorilla/websocket"
)

/* ======= Websockets connection pool ======= */

var AddConn (chan *websocket.Conn) = make(chan *websocket.Conn)
var DelConn (chan *websocket.Conn) = make(chan *websocket.Conn)
var SendString (chan string) = make(chan string)
var SendJSON (chan interface{}) = make(chan interface{})

func StartWebsocketPool() {
	conns := make(map[*websocket.Conn]bool)
	prevSend := int64(0)

	printIfNotCausedBySend := func(a ...interface{}) {
		if time.Now().Unix()-prevSend > 5 {
			print.Line(a...)
		} else {
			print.Debug(a...)
		}
	}

	for {
		select {
		case conn := <-AddConn:
			conns[conn] = true
			printIfNotCausedBySend("Got new connection. Total number:", cyan(len(conns)))

		case conn := <-DelConn:
			delete(conns, conn)
			printIfNotCausedBySend("Removed connection. Total number:", cyan(len(conns)))

		case msg := <-SendString:
			prevSend = time.Now().Unix()

			for conn, _ := range conns {
				err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
				if err != nil {
					print.Line("Error while writing:")
					print.Line(err)
					conn.Close()
					DelConn <- conn
				}
			}

		case data := <-SendJSON:
			prevSend = time.Now().Unix()

			dataString, _ := json.Marshal(data)
			print.Debug("sending:", string(dataString))

			for conn, _ := range conns {
				err := conn.WriteJSON(data)

				if err != nil {
					print.Line("Error sending json:", err)
					conn.Close()
					DelConn <- conn
				}
			}

		}
	}
}
