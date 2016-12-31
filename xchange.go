package sockets

import "errors"

type WsPack struct {
	ID  string
	Msg WsMsg
}

var sendChan chan *WsPack = make(chan *WsPack, 10)

func Broadcast(msg WsMsg) {
	pack := &WsPack{"*", msg}
	sendChan <- pack
}

func BroadcastData(topic string, content string) {
	Broadcast(WsMsg{topic, content})
}

func Send(id string, category string, msg string) {
	pack := &WsPack{id, WsMsg{Type: category, Payload: msg}}
	sendChan <- pack

}

func writeMessageToSocket(client SocketClient, msg WsMsg) (err error) {
	defer func() {
		if recover() != nil {
			err = errors.New("Dead socket")
		}
	}()
	return client.socket.WriteJSON(msg)
}

func serveWS() {
	go func() {
		for {
			pack := <-sendChan
			id := pack.ID
			msg := pack.Msg
			isBroadcast := id == "*"

			var deadConnections = 0
			var totalDeadConnections = 0
			toRemove := ClientIDs{}

			socketClients.RLock()
			for e := socketClients.clients.Front(); e != nil; e = e.Next() {
				if !e.Value.(SocketClient).isAlive {
					totalDeadConnections++
				} else if isBroadcast || (e.Value.(SocketClient).id == id) {
					err := writeMessageToSocket(e.Value.(SocketClient), msg)
					if err != nil {
						toRemove[e.Value.(SocketClient).id] = true
						deadConnections++
					}
				}
			}
			socketClients.RUnlock()
			if totalDeadConnections+deadConnections > 2 {
				removeClientsByID(toRemove)
				cleanDeadSockets()
			}

		}
	}()
}
