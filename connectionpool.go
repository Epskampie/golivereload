package main

import (
	"encoding/json"
	"livereload/print"

	"github.com/gorilla/websocket"
)

/* ======= Websockets connection pool ======= */

var AddConn (chan *websocket.Conn) = make(chan *websocket.Conn)
var DelConn (chan *websocket.Conn) = make(chan *websocket.Conn)
var SendString (chan string) = make(chan string)
var SendJSON (chan interface{}) = make(chan interface{})

func StartWebsocketPool() {
	print.Line("Starting connection pool")
	conns := make(map[*websocket.Conn]bool)

	for {
		select {
		case conn := <-AddConn:
			conns[conn] = true
			print.Line("Got new connection. Total number:", cyan(len(conns)))

		case conn := <-DelConn:
			delete(conns, conn)
			print.Line("Removed connection. Total number:", cyan(len(conns)))

		case msg := <-SendString:
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
			for conn, _ := range conns {
				dataString, err := json.Marshal(data)

				print.Line(string(dataString), err)
				err = conn.WriteJSON(data)

				if err != nil {
					print.Line("Error sending json:")
					print.Line(err)
				} else {
					print.Line("Sent write to")
				}
			}

		}
	}
}
