package handler

import (
	"context"
	"log"
	"net/http"
	"server/db"
	"server/ws"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}

	ws.GetHub().Register(conn)
	defer ws.GetHub().Unregister(conn)

	for {
		var msg ws.Message
		if err := conn.ReadJSON(&msg); err != nil {
			log.Println("read:", err)
			return
		}

		msg.Time = time.Now().UTC()

		_, err := db.Client.Message.Create().
			SetName(msg.Name).
			SetMessage(msg.Message).
			SetCreatedAt(msg.Time).
			Save(context.Background())
		if err != nil {
			log.Println("save:", err)
			continue
		}

		ws.GetHub().Broadcast(msg)
	}
}
