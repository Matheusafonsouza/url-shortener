FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/urlshortener ./cmd/server

FROM alpine:3.20

WORKDIR /app

RUN adduser -D -g '' appuser

COPY --from=builder /bin/urlshortener /app/urlshortener
COPY web/templates /app/web/templates

USER appuser

EXPOSE 8080

CMD ["/app/urlshortener"]
