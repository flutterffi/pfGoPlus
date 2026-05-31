FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /bin/pfgo-plus ./cmd/server

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates sqlite

COPY --from=builder /bin/pfgo-plus /app/pfgo-plus
COPY configs /app/configs

RUN mkdir -p /app/tmp

EXPOSE 8080

CMD ["/app/pfgo-plus"]
