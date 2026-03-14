FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/server \
    ./cmd/server

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S app -G app

WORKDIR /app

COPY --from=builder /out/server /app/server
# 不复制 config.yaml: 该文件含密钥且已被 .gitignore 排除,
# 运行配置通过环境变量 GO_BLOG_* 注入 (见 deploy/docker-compose.yml)。
COPY web /app/web

RUN mkdir -p /app/uploads/articles /app/uploads/avatars \
    && chown -R app:app /app/uploads

USER app
EXPOSE 8080

ENTRYPOINT ["/app/server"]
