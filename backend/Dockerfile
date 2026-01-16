FROM golang:1.21-alpine AS builder

WORKDIR /src
RUN apk add --no-cache build-base sqlite-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /out/dblive-server ./cmd/server

FROM alpine:3.19
WORKDIR /app
RUN apk add --no-cache sqlite-libs ca-certificates

COPY --from=builder /out/dblive-server ./dblive-server
COPY sqlite ./sqlite

ENV DBL_SERVER_ADDR=:8080
ENV DBL_DB_PATH=/app/sqlite/dblive.db

EXPOSE 8080
CMD ["./dblive-server"]
