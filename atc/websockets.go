package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func handleIO(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	ioJson.RLock()
	enc := json.NewEncoder(w)
	enc.Encode(ioJson)
	ioJson.RUnlock()
}

var ioConnections map[*websocket.Conn]bool
var manConnections map[*websocket.Conn]bool
var loggerConnections map[*websocket.Conn]bool
var Commands chan Command

type Command struct {
	Cmd       string `json:"cmd"`
	Parameter string `json:"parm"`
}

type wsLoggerMessage struct {
	Msg     string `json:"msg"`
	Context string `json:"context"`
}

func sendIO(msg []byte) {
	for conn := range ioConnections {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			delete(ioConnections, conn)
			conn.Close()
		}
	}
}

func sendMan(msg []byte) {
	for conn := range manConnections {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			delete(manConnections, conn)
			conn.Close()
		}
	}
}

func sendLog(msg string, context string) {
	json, err := json.Marshal(wsLoggerMessage{Msg: msg, Context: context})
	if err != nil {
		fmt.Println(err)
	}
	for conn := range loggerConnections {
		if err := conn.WriteMessage(websocket.TextMessage, json); err != nil {
			delete(loggerConnections, conn)
			conn.Close()
		}
	}
}

func IOWebSocketsHandler(w http.ResponseWriter, r *http.Request) {
	// Taken from gorilla's website
	conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		log.Println(err)
		return
	}
	log.Println("Succesfully upgraded connection")
	ioConnections[conn] = true

	for {
		// Blocks until a message is read
		_, msg, err := conn.ReadMessage()
		if err != nil {
			delete(ioConnections, conn)
			log.Println("Connection Closed")
			conn.Close()
			return
		}
		log.Println(string(msg))
		// sendIO(msg)
	}
}

func ManualWebSocketsHandler(w http.ResponseWriter, r *http.Request) {
	// Taken from gorilla's website
	conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		log.Println(err)
		return
	}
	log.Println("Succesfully upgraded connection")
	manConnections[conn] = true
	b, err := json.Marshal(atcControls)
	if err != nil {
		fmt.Println(err)
	}
	sendMan(b)
	for {
		// Blocks until a message is read
		_, msg, err := conn.ReadMessage()
		if err != nil {
			delete(manConnections, conn)
			log.Println("Connection Closed")
			conn.Close()
			return
		}
		var command Command
		json.Unmarshal(msg, command)
		Commands <- command
	}
}

func LOGWebSocketsHandler(w http.ResponseWriter, r *http.Request) {
	// Taken from gorilla's website
	conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		log.Println(err)
		return
	}
	log.Println("Succesfully upgraded connection")
	loggerConnections[conn] = true
	// b, err := json.Marshal(atcControls)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// sendLog(b)
	for {
		// Blocks until a message is read
		_, msg, err := conn.ReadMessage()
		if err != nil {
			delete(loggerConnections, conn)
			log.Println("Connection Closed")
			conn.Close()
			return
		}
		log.Println("From Logger:", msg, " ", conn.RemoteAddr())
	}
}

func JsonEncode(msg ioStatus) []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
	}
	return b
}
