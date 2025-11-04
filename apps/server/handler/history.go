package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"server/db"
	"server/ent"
	"server/ent/message"
	"sort"
	"time"
)

type MessageResponse struct {
	Name    string    `json:"name"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

func HandleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	msgs, err := db.Client.Message.Query().
		Order(ent.Desc(message.FieldCreatedAt)).
		Limit(10).
		All(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 古い順にソート
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].CreatedAt.Before(msgs[j].CreatedAt)
	})

	var result []MessageResponse
	for _, m := range msgs {
		result = append(result, MessageResponse{
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
