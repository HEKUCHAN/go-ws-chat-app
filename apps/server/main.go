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

	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/ws", handler.HandleWebSocket)
	apiMux.HandleFunc("/history", handler.HandleHistory)

	// /api/ 以下としてマウント
	http.Handle("/api/", http.StripPrefix("/api", apiMux))

	// サーバー起動
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
