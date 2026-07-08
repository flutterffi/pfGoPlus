FROM golang:1.24-alpine AS builder

ARG ENTRYPOINT=./cmd/server
ARG BIN_NAME=pfgo-plus

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /bin/${BIN_NAME} ${ENTRYPOINT}

FROM alpine:3.22

ARG BIN_NAME=pfgo-plus

WORKDIR /app

RUN apk add --no-cache ca-certificates sqlite

COPY --from=builder /bin/${BIN_NAME} /app/${BIN_NAME}
COPY configs /app/configs

RUN mkdir -p /app/tmp

EXPOSE 8080

CMD ["/bin/sh", "-c", "/app/${BIN_NAME}"]
