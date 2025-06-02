FROM golang:1.24.3 AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y make

COPY . .

RUN make build-controller

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates

WORKDIR /app

COPY --from=builder /app/bin/controller /app/controller


ENTRYPOINT ["/app/controller"]