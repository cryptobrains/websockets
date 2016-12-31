package sockets

import (
	"fmt"
	"github.com/leavengood/websocket"
	"github.com/valyala/fasthttp"
)

type WsMsg struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

type WSCtrl struct {
	ID      string          `json:"id"`
	IDS     map[string]bool `json:"ids"`
	Command string          `json:"command"`
}

type HandleWSMessage func(command string, id string, v interface{})

var websocketUpgrader *websocket.FastHTTPUpgrader
var WSMessageHandler HandleWSMessage = nil

func printAllConnections() {
	connections.RLock()
	for id := range connections.allConns {
		fmt.Println(id)
	}
	connections.RUnlock()
}

func in(ws *websocket.Conn) {

	for {
		in := WSCtrl{}

		err := ws.ReadJSON(&in)
		if err != nil {
			if _, ok := err.(*websocket.CloseError); ok {
				removeBySocket(ws) //TODOv2: Queue these
			}
			ws.Close()
			break
		}

		command := in.Command
		id := in.ID

		switch command {
		case "ADD":
			addClient(ws, id)

		case "LOGOUT":
			removeClientByID(id)

		}
		go WSMessageHandler(command, id, in)
	}
}

func HandleWS(ctx *fasthttp.RequestCtx) {
	websocketUpgrader.UpgradeHandler(ctx)
}

func Init() {

	websocketUpgrader = &websocket.FastHTTPUpgrader{
		CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
			return true
		},
		Handler: in,
	}
	serveWS()
}
