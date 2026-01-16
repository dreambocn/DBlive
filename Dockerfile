FROM node:20-alpine AS frontend-builder
WORKDIR /app

COPY frontend/package.json ./
RUN npm install

COPY frontend/ ./frontend/
WORKDIR /app/frontend
RUN npm run build

FROM golang:1.21-alpine AS backend-builder
WORKDIR /src
RUN apk add --no-cache build-base sqlite-dev

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /out/dblive-server ./cmd/server

FROM alpine:3.19
WORKDIR /app
RUN apk add --no-cache sqlite-libs ca-certificates

COPY --from=backend-builder /out/dblive-server ./dblive-server
COPY --from=frontend-builder /app/frontend/dist ./public
COPY backend/sqlite ./sqlite

ENV DBL_SERVER_ADDR=:8080
ENV DBL_DB_PATH=/app/sqlite/dblive.db

EXPOSE 8080
CMD ["./dblive-server"]
