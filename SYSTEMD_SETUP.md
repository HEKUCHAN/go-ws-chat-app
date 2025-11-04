# systemd サービスとして登録する手順

このアプリケーションをLinuxサーバーで自動起動させるための手順です。

## 前提条件

- Docker と Docker Compose がインストールされていること
- systemd を使用しているLinux環境であること（Ubuntu, CentOS, Debian等）

## 手順

### 1. サービスファイルを配置

```bash
# プロジェクトのパスを確認（このREADME.mdがある場所）
pwd

# サービスファイルを編集して正しいパスに変更
sudo nano /path/to/go-ws-chat-app/docker-compose-app.service

# WorkingDirectoryを実際のプロジェクトパスに変更してください
# 例: WorkingDirectory=/home/ubuntu/go-ws-chat-app
```

### 2. サービスファイルをsystemdディレクトリにコピー

```bash
sudo cp docker-compose-app.service /etc/systemd/system/
```

### 3. 権限を設定

```bash
sudo chmod 644 /etc/systemd/system/docker-compose-app.service
```

### 4. systemdをリロード

```bash
sudo systemctl daemon-reload
```

### 5. サービスを有効化（起動時に自動起動）

```bash
sudo systemctl enable docker-compose-app.service
```

### 6. サービスを開始

```bash
sudo systemctl start docker-compose-app.service
```

### 7. 状態を確認

```bash
sudo systemctl status docker-compose-app.service
```

## サービスの管理コマンド

```bash
# サービスを開始
sudo systemctl start docker-compose-app.service

# サービスを停止
sudo systemctl stop docker-compose-app.service

# サービスを再起動
sudo systemctl restart docker-compose-app.service

# サービスの状態を確認
sudo systemctl status docker-compose-app.service

# サービスのログを表示
sudo journalctl -u docker-compose-app.service -f

# 自動起動を無効化
sudo systemctl disable docker-compose-app.service

# 自動起動を有効化
sudo systemctl enable docker-compose-app.service
```

## トラブルシューティング

### Dockerコマンドのパスを確認

```bash
which docker
# 通常は /usr/bin/docker
```

もし違うパスの場合は、サービスファイルの `ExecStart` と `ExecStop` を修正してください。

### Docker Composeのバージョン確認

```bash
docker compose version
```

古いバージョンの場合は `docker-compose` (ハイフン付き) を使用してください：

```bash
# サービスファイル内を以下に変更
ExecStart=/usr/bin/docker-compose up -d
ExecStop=/usr/bin/docker-compose down
```

### サービスが起動しない場合

```bash
# 詳細なログを確認
sudo journalctl -xe -u docker-compose-app.service

# サービスファイルの構文チェック
sudo systemd-analyze verify /etc/systemd/system/docker-compose-app.service
```

## 注意事項

- サーバー再起動後、Dockerサービスが起動してからこのサービスが起動します
- `restart: unless-stopped` が docker-compose.yml に設定されているため、コンテナは自動的に再起動されます
- 手動で `docker compose down` を実行した場合、サービスは停止状態のままになります
