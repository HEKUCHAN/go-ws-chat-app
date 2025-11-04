package main

import (
	"log"
	"net/http"
	"server/db"
	"server/handler"
)

func main() {
	// データベース初期化
	db.Init()

	// ルーティング設定
	http.HandleFunc("/ws", handler.HandleWebSocket)
	http.HandleFunc("/history", handler.HandleHistory)

	// サーバー起動
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
