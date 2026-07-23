FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates gcc musl-dev

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /sniper ./cmd/main.go

# ── Runtime stage ──────────────────────────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /sniper /sniper

EXPOSE 8080

ENTRYPOINT ["/sniper"]
