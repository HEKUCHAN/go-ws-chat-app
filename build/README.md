# ビルド方法

## Dockerイメージのビルド

リポジトリルートから実行：

```bash
cd <Some Path>/go-ws-chat-app # change directory to project root
docker build --platform=linux/amd64 -f build/Dockerfile -t go-ws-chat-server:latest .
```

## バイナリの取り出し

コンテナ化せずに生のバイナリだけが欲しい場合：

```bash
# 一時コンテナからバイナリを取り出す
cid=$(docker create go-ws-chat-server:latest)
docker cp "$cid":/usr/local/bin/app ./apps/server/app_linux_amd64
docker rm "$cid"
```

## 実行

### Dockerコンテナで実行

```bash
docker run --rm -p 8080:8080 go-ws-chat-server:latest
```

### バイナリを直接実行（Linux環境）

インストールする側で必要な依存関係：

```bash
sudo apt-get update && sudo apt-get install -y libsqlite3-0
```

実行：

```bash
./app-linux-amd64
```