FROM golang:1.24-alpine AS builder
WORKDIR /src

# Cache modules
COPY go.mod go.sum ./
RUN apk add --no-cache git && \
    go env -w GOPROXY=https://proxy.golang.org && \
    go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/server ./server

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
