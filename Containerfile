FROM golang:1.21-alpine AS builder

WORKDIR /build

COPY go.mod go.sum* ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o blacklistupdater ./cmd/blacklistupdater

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/blacklistupdater /app/blacklistupdater
COPY config.yaml /app/config.yaml

VOLUME ["/app/data"]

ENTRYPOINT ["/app/blacklistupdater"]
CMD ["-config", "/app/config.yaml"]
