package ws

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Name    string    `json:"name"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

type ConnectionHub struct {
	clients   map[*websocket.Conn]struct{}
	mu        sync.RWMutex
	broadcast chan Message
}

var hub *ConnectionHub

func init() {
	hub = newHub()
}

func GetHub() *ConnectionHub {
	return hub
}

func newHub() *ConnectionHub {
	h := &ConnectionHub{
		clients:   make(map[*websocket.Conn]struct{}),
		broadcast: make(chan Message, 128),
	}
	go h.run()
	return h
}

func (h *ConnectionHub) Register(conn *websocket.Conn) {
	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()
}

func (h *ConnectionHub) Unregister(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, conn)
	conn.Close()
	h.mu.Unlock()
}

func (h *ConnectionHub) Broadcast(msg Message) {
	h.broadcast <- msg
}

func (h *ConnectionHub) run() {
	for msg := range h.broadcast {
		h.mu.RLock()
		for client := range h.clients {
			if err := client.WriteJSON(msg); err != nil {
				continue
			}
		}
		h.mu.RUnlock()
	}
}
