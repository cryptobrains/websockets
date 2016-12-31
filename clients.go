package sockets

import (
	"container/list"
	"github.com/leavengood/websocket"
	"sync"
)

type SocketClient struct {
	id      string
	isAlive bool
	socket  *websocket.Conn
}

type ClientIDs map[string]bool

var socketClients = struct {
	sync.RWMutex
	clients *list.List
}{clients: list.New()}

var connections = struct {
	sync.RWMutex
	allClients map[string]bool
	allConns   map[string]*websocket.Conn
}{allClients: make(map[string]bool), allConns: make(map[string]*websocket.Conn)}

func disabledClient(client SocketClient) SocketClient {
	client.socket.Close()
	return SocketClient{id: client.id, isAlive: false, socket: client.socket}
}

func removeClientByID(id string) {
	socketClients.Lock()
	defer socketClients.Unlock()
	for e := socketClients.clients.Front(); e != nil; e = e.Next() {
		if e.Value.(SocketClient).id == id {
			e.Value = disabledClient(e.Value.(SocketClient))
			break
		}
	}
}

func removeClientsByID(ids ClientIDs) {
	socketClients.Lock()
	defer socketClients.Unlock()
	for e := socketClients.clients.Front(); e != nil; e = e.Next() {
		_, ok := ids[e.Value.(SocketClient).id]
		if ok {
			e.Value = disabledClient(e.Value.(SocketClient))
		}
	}
}

func removeBySocket(ws *websocket.Conn) {
	socketClients.Lock()
	defer socketClients.Unlock()
	for e := socketClients.clients.Front(); e != nil; e = e.Next() {
		if e.Value.(SocketClient).socket == ws {
			e.Value = disabledClient(e.Value.(SocketClient))
			break
		}
	}
}

func markEveryoneAsDisconnected() {
	socketClients.Lock()
	defer socketClients.Unlock()
	for e := socketClients.clients.Front(); e != nil; e = e.Next() {
		e.Value = disabledClient(e.Value.(SocketClient))

	}
}

func cleanDeadSockets() {
	socketClients.Lock()
	defer socketClients.Unlock()
	item := socketClients.clients.Front()
	for {
		nextItem := item.Next()
		if !item.Value.(SocketClient).isAlive {
			item.Value.(SocketClient).socket.Close()
			socketClients.clients.Remove(item)
		}
		if nextItem == nil {
			break
		} else {
			item = nextItem
		}

	}
}

func DisconnectEveryone() {
	markEveryoneAsDisconnected()
	cleanDeadSockets()
}

func CountClients() int {
	var retVal = 0
	socketClients.RLock()
	defer socketClients.RUnlock()
	for e := socketClients.clients.Front(); e != nil; e = e.Next() {
		if e.Value.(SocketClient).isAlive {
			retVal++
		}
	}
	return retVal
}

func addClient(conn *websocket.Conn, id string) bool {
	var existing = false
	socketClients.RLock()
	for e := socketClients.clients.Front(); e != nil; e = e.Next() {
		if (e.Value.(SocketClient).id == id) && (e.Value.(SocketClient).isAlive) {
			existing = true
			break
		}

	}
	socketClients.RUnlock()
	if existing {
		return false
	}
	// We may still have duplicates, but for the most part they will self-manage. In some cases it will be an unrecorded disconnection and teh earlier connectionw ill be killed. In the very unlikely worse case, the duplicate will be ignored.

	socketClients.Lock()
	defer socketClients.Unlock()

	item := SocketClient{id: id, isAlive: true, socket: conn}
	socketClients.clients.PushFront(item)

	return true
}
