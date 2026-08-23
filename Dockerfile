# Stage 1: Use a Go image to build the application
FROM golang:1.26.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6
RUN apk add --no-cache postgresql-client

COPY . .
# Generate Swagger package if missing (e.g. fresh clone); no-op when docs/docs.go exists
RUN swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api

# Stage 2: Minimal runtime image
FROM alpine:3.18

RUN apk --no-cache add ca-certificates postgresql-client

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/database/migrations ./database/migrations/
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

RUN mkdir -p logs

COPY scripts/entrypoint.sh /entrypoint.sh
COPY scripts/parse-database-url.sh /scripts/parse-database-url.sh
COPY scripts/parse-redis-url.sh /scripts/parse-redis-url.sh
RUN chmod +x /entrypoint.sh

EXPOSE ${PORT}

ENTRYPOINT ["/entrypoint.sh"]
CMD ["./main"]
