package main

import (
	"html"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

var (
	clients   = make(map[*websocket.Conn]struct{})
	clientsMu sync.RWMutex
	broadcast = make(chan Message, 128)
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func init() {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
}

func main() {
	go broadcaster()

	http.HandleFunc("/ws", handleWS)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	register(conn)
	defer unregister(conn)

	conn.SetReadLimit(1 << 20)
	resetReadDeadline := func() { conn.SetReadDeadline(time.Now().Add(60 * time.Second)) }
	resetReadDeadline()
	conn.SetPongHandler(func(string) error { resetReadDeadline(); return nil })

	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("read:", err)
			}
			return
		}
		msg.Name = sanitize(msg.Name, 32)
		msg.Message = sanitize(msg.Message, 512)
		if msg.Name == "" && msg.Message == "" {
			continue
		}
		broadcast <- msg
	}
}

func broadcaster() {
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case msg := <-broadcast:
			clientsMu.RLock()
			for c := range clients {
				c.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := c.WriteJSON(msg); err != nil {
					clientsMu.RUnlock()
					unregister(c)
					clientsMu.RLock()
				}
			}
			clientsMu.RUnlock()
		case <-ticker.C:
			clientsMu.RLock()
			for c := range clients {
				c.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := c.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(10*time.Second)); err != nil {
					clientsMu.RUnlock()
					unregister(c)
					clientsMu.RLock()
				}
			}
			clientsMu.RUnlock()
		}
	}
}

func sanitize(s string, max int) string {
	s = strings.TrimSpace(s)
	s = strings.Map(func(r rune) rune {
		switch r {
		case '\n', '\r', '\t':
			return ' '
		}
		if r < 0x20 {
			return -1
		}
		return r
	}, s)
	runes := []rune(s)
	if len(runes) > max {
		s = string(runes[:max])
	}
	return html.EscapeString(s)
}

func register(c *websocket.Conn) {
	clientsMu.Lock()
	clients[c] = struct{}{}
	clientsMu.Unlock()
}

func unregister(c *websocket.Conn) {
	clientsMu.Lock()
	if _, ok := clients[c]; ok {
		delete(clients, c)
		_ = c.Close()
	}
	clientsMu.Unlock()
}
