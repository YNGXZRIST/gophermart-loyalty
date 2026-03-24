FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /gophermart ./cmd/gophermart

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /gophermart /usr/local/bin/gophermart

ENV RUN_ADDRESS=:8080 \
    DATABASE_URI=postgres://gophermart:gophermart@db:5432/gophermart?sslmode=disable \
    ACCRUAL_SYSTEM_ADDRESS=http://accrual:8080 \
    MODE=release \
    ACCRUAL_WORKER_COUNT=4 \
    ACCRUAL_USE_MOCK=false

EXPOSE 8080

CMD ["gophermart"]
