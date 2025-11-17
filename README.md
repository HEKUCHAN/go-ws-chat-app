# Go WebSocket Chat — jQuery版 (Single Shared Room)

同じ1ルームチャットを **jQuery** を使ってクライアント実装する版です。

---

## ファイル構成

```
./
├─ go.mod
├─ main.go
└─ static/
   └─ index.html
```

---

## go.mod

```go
module github.com/example/ws-chat

go 1.22.0

require github.com/gorilla/websocket v1.5.1
```
