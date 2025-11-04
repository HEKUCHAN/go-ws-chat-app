package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"server/db"
	"server/ent"
	"server/ent/message"
	"sort"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type MessagePayload struct {
	Name    string    `json:"name"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

var (
	ClientsMu sync.RWMutex
	Clients   = make(map[*websocket.Conn]struct{})
	Broadcast = make(chan MessagePayload, 128)
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	register(conn)
	defer unregister(conn)

	for {
		var msg MessagePayload
		if err := conn.ReadJSON(&msg); err != nil {
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
		Broadcast <- msg
	}
}

func register(c *websocket.Conn) {
	ClientsMu.Lock()
	Clients[c] = struct{}{}
	ClientsMu.Unlock()
}

func unregister(c *websocket.Conn) {
	ClientsMu.Lock()
	delete(Clients, c)
	_ = c.Close()
	ClientsMu.Unlock()
}

func sendRecentMessages(conn *websocket.Conn, n int) error {
	ctx := context.Background()
	msgs, err := db.Client.Message.Query().
		Order(ent.Desc(message.FieldCreatedAt)).
		Limit(n).
		All(ctx)
	if err != nil {
		return err
	}
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].CreatedAt.Before(msgs[j].CreatedAt)
	})
	for _, m := range msgs {
		payload := MessagePayload{
			Name:    m.Name,
			Message: m.Message,
			Time:    m.CreatedAt,
		}
		if err := conn.WriteJSON(payload); err != nil {
			return err
		}
	}
	return nil
}

func HandleHistory(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	msgs, err := db.Client.Message.Query().
		Order(ent.Desc(message.FieldCreatedAt)).
		Limit(10).
		All(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].CreatedAt.Before(msgs[j].CreatedAt)
	})

	var result []MessagePayload
	for _, m := range msgs {
		result = append(result, MessagePayload{
			Name:    m.Name,
			Message: m.Message,
			Time:    m.CreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
