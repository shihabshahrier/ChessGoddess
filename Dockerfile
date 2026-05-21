FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /chesslens ./cmd/server

FROM alpine:3.21

RUN apk add --no-cache ca-certificates stockfish

COPY --from=builder /chesslens /usr/local/bin/chesslens

EXPOSE 8080

ENTRYPOINT ["chesslens"]
