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

---

## main.go

```go
package main

import (
	"html"
	"log"
	"net/http"
	"net/url"
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
	// 本番向け: 同一ホストからの Origin のみ許可（ローカル開発では Origin 無しを許可）
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" { // curl やファイル直開き等
			return true
		}
		u, err := url.Parse(origin)
		if err != nil {
			return false
		}
		// Host 一致のみ判定（必要に応じてスキームやポートを厳格化）
		return strings.EqualFold(u.Host, r.Host)
	},
}

func main() {
	go broadcaster()

	// セキュアな静的配信（最低限のヘッダを付与）
	http.Handle("/static/", http.StripPrefix("/static/", secureHeaders(http.FileServer(http.Dir("static")))))
	// WebSocket エンドポイント
	http.HandleFunc("/ws", handleWS)

	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
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
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("read:", err)
			}
			return
		}
		// ---- XSS/入力対策（サーバ側） ----
		msg.Name = sanitize(msg.Name, 32)
		msg.Message = sanitize(msg.Message, 512)
		if msg.Name == "" && msg.Message == "" {
			continue
		}
		broadcast <- msg
	}
}

// sanitize はトリミング・長さ制限・改行/制御文字の除去・HTMLエスケープを行う
func sanitize(s string, max int) string {
	s = strings.TrimSpace(s)
	// 制御文字類を除去/正規化
	s = strings.Map(func(r rune) rune {
		if r == '
' || r == '
' || r == '	' {
			return ' '
		}
		if r < 0x20 { // その他制御文字
			return -1
		}
		return r
	}, s)
	// 文字数制限（rune 単位）
	if len([]rune(s)) > max {
		runes := []rune(s)
		s = string(runes[:max])
	}
	// HTML エスケープ（二重化防御：クライアントでも textNode で描画）
	return html.EscapeString(s)
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

// 最低限のセキュリティヘッダ（CSPは簡略版）
func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' https://code.jquery.com; connect-src 'self' ws: wss:; style-src 'self' 'unsafe-inline'; object-src 'none'; base-uri 'none'")
		next.ServeHTTP(w, r)
	})
}
```

go
package main

import (
"html"
"log"
"net/http"
"strings"
"sync"
"time"

```
"github.com/gorilla/websocket"
```

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
CheckOrigin:     func(r *http.Request) bool { return true },
}

func main() {
go broadcaster()

```
// セキュアな静的配信（簡易ヘッダ付与）
http.Handle("/static/", http.StripPrefix("/static/", secureHeaders(http.FileServer(http.Dir("static")))))
// WebSocket エンドポイント
http.HandleFunc("/ws", handleWS)

log.Println("listening on :8080")
if err := http.ListenAndServe(":8080", nil); err != nil {
	log.Fatal(err)
}
```

}

func handleWS(w http.ResponseWriter, r *http.Request) {
conn, err := upgrader.Upgrade(w, r, nil)
if err != nil {
log.Println("upgrade:", err)
return
}
register(conn)
defer unregister(conn)

```
conn.SetReadLimit(1 << 20)
conn.SetReadDeadline(time.Now().Add(60 * time.Second))
conn.SetPongHandler(func(string) error {
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	return nil
})

for {
	var msg Message
	if err := conn.ReadJSON(&msg); err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			log.Println("read:", err)
		}
		return
	}
	// ---- XSS/入力対策（サーバ側） ----
	msg.Name = sanitize(msg.Name, 32)
	msg.Message = sanitize(msg.Message, 512)
	if msg.Name == "" && msg.Message == "" {
		continue
	}
	broadcast <- msg
}
```

}

// sanitize はトリミング・長さ制限・改行/制御文字の除去・HTMLエスケープを行う
func sanitize(s string, max int) string {
s = strings.TrimSpace(s)
// 制御文字類を空白に
s = strings.Map(func(r rune) rune {
if r == '
' || r == '
' || r == '	' {
return ' '
}
if r < 0x20 { // その他制御文字
return -1
}
return r
}, s)
if len([]rune(s)) > max {
runes := []rune(s)
s = string(runes[:max])
}
// HTML エスケープ（防御の二重化：クライアントでもエスケープ）
return html.EscapeString(s)
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

// 最低限のセキュリティヘッダ（CSPは簡略版）
func secureHeaders(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.Header().Set("X-Content-Type-Options", "nosniff")
w.Header().Set("X-Frame-Options", "DENY")
w.Header().Set("Referrer-Policy", "no-referrer")
w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' [https://code.jquery.com](https://code.jquery.com); connect-src 'self' ws: wss:; style-src 'self' 'unsafe-inline'; object-src 'none'; base-uri 'none'")
next.ServeHTTP(w, r)
})
}

```go
package main

import (
	"log"
	"net/http"
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
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func main() {
	go broadcaster()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/ws", handleWS)

	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
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
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("read:", err)
			}
			return
		}
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
```

---

## static/index.html

```html
<!doctype html>
<html lang="ja">
<head>
  <meta charset="utf-8" />
  <meta http-equiv="X-UA-Compatible" content="IE=edge" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>jQuery WS Chat (XSS Hardened)</title>
  <script src="https://code.jquery.com/jquery-3.7.1.min.js" integrity="sha256-/JqT3SQfawRcv/BIHPThkBvs0OEvtFFmqPF/lYI/Cxo=" crossorigin="anonymous"></script>
  <style>
    body { font-family: system-ui, sans-serif; margin: 24px; }
    #log { border: 1px solid #ccc; padding: 12px; height: 300px; overflow-y: auto; }
    .msg { margin: 6px 0; }
    .name { font-weight: bold; margin-right: 6px; }
    form { display: flex; gap: 8px; margin-top: 12px; }
    input[type=text] { flex: 1; padding: 8px; }
    input[name=name] { width: 160px; }
  </style>
</head>
<body>
  <h1>WebSocket Chat — jQuery版（XSS対策・Origin検証）</h1>
  <div id="log" aria-live="polite" aria-atomic="false"></div>

  <form id="chatForm" autocomplete="off">
    <input type="text" name="name" placeholder="名前" maxlength="32" required />
    <input type="text" name="message" placeholder="メッセージ" maxlength="512" required />
    <button type="submit">送信</button>
  </form>

  <script>
    const wsScheme = location.protocol === 'https:' ? 'wss' : 'ws';
    const ws = new WebSocket(`${wsScheme}://${location.host}/ws`);

    ws.addEventListener('open', () => appendMsg('** 接続しました **'));
    ws.addEventListener('close', () => appendMsg('** 切断されました **'));
    ws.addEventListener('error', () => appendMsg('** エラーが発生しました **'));

    ws.addEventListener('message', (e) => {
      try {
        const data = JSON.parse(e.data);
        const line = document.createElement('div');
        line.className = 'msg';

        const name = document.createElement('span');
        name.className = 'name';
        name.appendChild(document.createTextNode(data.name || ''));

        const body = document.createElement('span');
        body.appendChild(document.createTextNode(data.message || ''));

        line.appendChild(name);
        line.appendChild(body);
        document.getElementById('log').appendChild(line);
        document.getElementById('log').scrollTop = document.getElementById('log').scrollHeight;
      } catch {
        appendMsg('** 不正なメッセージ **');
      }
    });

    $('#chatForm').on('submit', function(e) {
      e.preventDefault();
      if (ws.readyState !== WebSocket.OPEN) return;
      const name = $('input[name=name]').val().trim();
      const message = $('input[name=message]').val().trim();
      if (!name || !message) return;

      const norm = (s, max) => s.replace(/[

	]/g, ' ').slice(0, max);
      const payload = { name: norm(name, 32), message: norm(message, 512) };

      ws.send(JSON.stringify(payload));
      $('input[name=message]').val('');
    });

    function appendMsg(text) {
      const div = document.createElement('div');
      div.className = 'msg';
      div.appendChild(document.createTextNode(text));
      document.getElementById('log').appendChild(div);
      document.getElementById('log').scrollTop = document.getElementById('log').scrollHeight;
    }
  </script>
</body>
</html>
```

html

<!doctype html>

<html lang="ja">
<head>
  <meta charset="utf-8" />
  <meta http-equiv="X-UA-Compatible" content="IE=edge" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>jQuery WS Chat (XSS Hardened)</title>
  <script src="https://code.jquery.com/jquery-3.7.1.min.js" integrity="sha256-/JqT3SQfawRcv/BIHPThkBvs0OEvtFFmqPF/lYI/Cxo=" crossorigin="anonymous"></script>
  <style>
    body { font-family: system-ui, sans-serif; margin: 24px; }
    #log { border: 1px solid #ccc; padding: 12px; height: 300px; overflow-y: auto; }
    .msg { margin: 6px 0; }
    .name { font-weight: bold; margin-right: 6px; }
    form { display: flex; gap: 8px; margin-top: 12px; }
    input[type=text] { flex: 1; padding: 8px; }
    input[name=name] { width: 160px; }
  </style>
</head>
<body>
  <h1>WebSocket Chat — jQuery版（XSS対策）</h1>
  <div id="log" aria-live="polite" aria-atomic="false"></div>

  <form id="chatForm" autocomplete="off">
    <input type="text" name="name" placeholder="名前" maxlength="32" required />
    <input type="text" name="message" placeholder="メッセージ" maxlength="512" required />
    <button type="submit">送信</button>
  </form>

  <script>
    const wsScheme = location.protocol === 'https:' ? 'wss' : 'ws';
    const ws = new WebSocket(`${wsScheme}://${location.host}/ws`);

    ws.addEventListener('open', () => appendMsg('** 接続しました **'));
    ws.addEventListener('close', () => appendMsg('** 切断されました **'));
    ws.addEventListener('error', () => appendMsg('** エラーが発生しました **'));

    ws.addEventListener('message', (e) => {
      // 受信は JSON を想定。サーバ側でも HTML エスケープ済み。
      try {
        const data = JSON.parse(e.data);
        // 念のためクライアント側でもテキストノードで追加（innerHTML 非使用）
        const line = document.createElement('div');
        line.className = 'msg';

        const name = document.createElement('span');
        name.className = 'name';
        name.appendChild(document.createTextNode(data.name || ''));

        const body = document.createElement('span');
        body.appendChild(document.createTextNode(data.message || ''));

        line.appendChild(name);
        line.appendChild(body);
        document.getElementById('log').appendChild(line);
        document.getElementById('log').scrollTop = document.getElementById('log').scrollHeight;
      } catch {
        appendMsg('** 不正なメッセージ **');
      }
    });

    $('#chatForm').on('submit', function(e) {
      e.preventDefault();
      if (ws.readyState !== WebSocket.OPEN) return;
      const name = $('input[name=name]').val().trim();
      const message = $('input[name=message]').val().trim();
      if (!name || !message) return;

      // 送信前にクライアント側でも長さ・改行を軽く正規化
      const norm = (s, max) => s.replace(/[

	]/g, ' ').slice(0, max);
      const payload = { name: norm(name, 32), message: norm(message, 512) };

      ws.send(JSON.stringify(payload));
      $('input[name=message]').val('');
    });

    function appendMsg(text) {
      const div = document.createElement('div');
      div.className = 'msg';
      div.appendChild(document.createTextNode(text));
      document.getElementById('log').appendChild(div);
      document.getElementById('log').scrollTop = document.getElementById('log').scrollHeight;
    }
  </script>

</body>
</html>
```
html
<!doctype html>
<html lang="ja">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>jQuery WS Chat</title>
  <script src="https://code.jquery.com/jquery-3.7.1.min.js"></script>
  <style>
    body { font-family: system-ui, sans-serif; margin: 24px; }
    #log { border: 1px solid #ccc; padding: 12px; height: 300px; overflow-y: auto; }
    .msg { margin: 6px 0; }
    .name { font-weight: bold; color: #0074d9; margin-right: 6px; }
    form { display: flex; gap: 8px; margin-top: 12px; }
    input[type=text] { flex: 1; padding: 8px; }
    input[name=name] { width: 160px; }
  </style>
</head>
<body>
  <h1>WebSocket Chat — jQuery版</h1>
  <div id="log"></div>

  <form id="chatForm">
    <input type="text" name="name" placeholder="名前" required />
    <input type="text" name="message" placeholder="メッセージ" required />
    <button type="submit">送信</button>
  </form>

  <script>
    const wsScheme = location.protocol === 'https:' ? 'wss' : 'ws';
    const ws = new WebSocket(`${wsScheme}://${location.host}/ws`);

    ws.onopen = () => appendMsg('** 接続しました **');
    ws.onclose = () => appendMsg('** 切断されました **');
    ws.onerror = () => appendMsg('** エラーが発生しました **');

    ws.onmessage = function(e) {
      try {
        const data = JSON.parse(e.data);
        appendMsg(`<span class='name'>${escapeHtml(data.name)}</span>${escapeHtml(data.message)}`);
      } catch {
        appendMsg(e.data);
      }
    };

    $("#chatForm").on('submit', function(e) {
      e.preventDefault();
      const name = $("input[name=name]").val().trim();
      const message = $("input[name=message]").val().trim();
      if (!name || !message || ws.readyState !== WebSocket.OPEN) return;
      ws.send(JSON.stringify({ name, message }));
      $("input[name=message]").val('');
    });

    function appendMsg(html) {
      const div = $('<div>').addClass('msg').html(html);
      $('#log').append(div);
      $('#log').scrollTop($('#log')[0].scrollHeight);
    }

    function escapeHtml(text) {
      return $('<div>').text(text).html();
    }
  </script>

</body>
</html>
```

---

## 特徴

* jQueryでDOM操作とイベント管理を簡潔化。
* HTMLエスケープ処理 (`escapeHtml`) を追加してXSS対策。
* UI・挙動は前バージョンと同じ（単一ルーム・全員ブロードキャスト）。
