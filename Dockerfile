FROM golang:1.23.7-bookworm AS builder

ARG GO111MODULE=on

WORKDIR /go/src/websocket

COPY ./ ./

RUN go mod download

RUN CGO_ENABLED=0 go build -o ./bin/websocket ./

FROM alpine:3.21 AS app

WORKDIR /app

COPY --from=builder /go/src/websocket/bin/websocket .

EXPOSE 80

ENTRYPOINT ["/app/websocket"]
