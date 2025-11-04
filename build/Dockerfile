# Dockerfile for server
FROM ubuntu:24.04 AS build

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates curl build-essential git pkg-config libsqlite3-dev \
 && rm -rf /var/lib/apt/lists/*

# Go を取得
ENV GO_VERSION=1.22.7
RUN curl -fsSL https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz \
 | tar -C /usr/local -xz
ENV PATH=/usr/local/go/bin:$PATH

WORKDIR /src
# ビルドコンテキストはリポジトリルートを想定
COPY apps/server/go.mod apps/server/go.sum ./
RUN go mod download

COPY apps/server/ ./

# Linux/amd64 用に cgo を有効化
ENV CGO_ENABLED=1 GOOS=linux GOARCH=amd64
RUN go build -tags "sqlite_omit_load_extension" -ldflags="-s -w" -o /out/app

# 実行用の極小イメージ（glibc ランタイムと libsqlite3 が必要）
FROM ubuntu:24.04
RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates libsqlite3-0 \
 && rm -rf /var/lib/apt/lists/*
COPY --from=build /out/app /usr/local/bin/app
ENTRYPOINT ["/usr/local/bin/app"]
