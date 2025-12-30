FROM golang:1.21-alpine AS builder

WORKDIR /build

COPY go.mod go.sum* ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o dnsblacklist ./cmd/dnsblacklist

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/dnsblacklist /app/dnsblacklist
COPY config.yaml /app/config.yaml

VOLUME ["/data"]

ENTRYPOINT ["/app/dnsblacklist"]
CMD ["-config", "/app/config.yaml"]
