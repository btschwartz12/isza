FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o app main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/app .
RUN apk add --no-cache tzdata python3 py3-pip \
    && pip3 install --break-system-packages --no-cache-dir instagrapi==2.1.2 pillow==11.0.0
