FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o bot ./cmd/bot

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/bot .
COPY --from=builder /app/telegram.yml .
COPY --from=builder /app/locales ./locales
COPY --from=builder /app/migrations ./migrations

CMD ["./bot"]
