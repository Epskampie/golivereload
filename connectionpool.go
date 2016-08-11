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
