# syntax=docker/dockerfile:1

# ---------------------------------------------------------------------------
# Stage 1: Build the frontend (React + Vite)
# ---------------------------------------------------------------------------
FROM node:22-alpine AS ui-builder

WORKDIR /app/UI

COPY UI/package.json UI/package-lock.json ./
RUN npm ci

COPY UI/ ./
RUN npm run build

# ---------------------------------------------------------------------------
# Stage 2: Build the Go binary
# ---------------------------------------------------------------------------
FROM golang:1.26-alpine AS go-builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /app/recho main.go

# ---------------------------------------------------------------------------
# Stage 3: Final runtime image
# ---------------------------------------------------------------------------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S recho && adduser -S recho -G recho

WORKDIR /app

COPY --from=go-builder /app/recho ./recho
COPY --from=go-builder /app/internal/infra/postgres/migrations ./internal/infra/postgres/migrations
COPY --from=ui-builder /app/UI/dist ./UI/dist

USER recho

EXPOSE 8000

ENTRYPOINT ["./recho"]
