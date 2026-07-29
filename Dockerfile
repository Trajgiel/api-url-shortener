FROM golang:1.26-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=1
RUN go build -ldflags="-w -s" -o /app/bin/url-shortener ./cmd/url-shortener


FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/bin/url-shortener ./url-shortener

RUN mkdir -p /app/storage

EXPOSE 8080

CMD ["sh", "-c", "HTTP_ADDRESS=0.0.0.0:${PORT:-8080} ./url-shortener"]
