# Stage 1: Build
FROM golang:1.26.0-bookworm AS builder

WORKDIR /platform

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go vet -v ./...
RUN go test -v ./...

# Build binaries
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/migrate ./cmd/migrate


# Stage 2: Final image
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y

# Копируем только бинарник из предыдущего этапа
COPY --from=builder /out/api /api
COPY --from=builder /out/migrate /migrate

USER 65532:65532