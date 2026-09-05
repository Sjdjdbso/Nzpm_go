# ── Stage 1: Build Binary Golang ──────────────────────────────────────────
FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o go-mirror-bot .

# ── Stage 2: Runtime Image Ringan (Hanya ~140MB) ──────────────────────────
FROM alpine:3.20

WORKDIR /app

# Install dependensi esensial (Aria2, Rclone, 7zip, ffmpeg, yt-dlp)
RUN apk add --no-cache \
    aria2 \
    rclone \
    ca-certificates \
    curl \
    bash \
    p7zip \
    ffmpeg \
    yt-dlp

# Salin binary dari stage builder & skrip runner
COPY --from=builder /app/go-mirror-bot .
COPY aria.sh start.sh ./
RUN chmod +x aria.sh start.sh

# Port default untuk HTTP Health Check Koyeb
EXPOSE 8080

# Jalankan via start.sh
CMD ["./start.sh"]
