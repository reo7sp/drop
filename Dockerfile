FROM golang:1.26.6 AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /usr/local/bin/app .


FROM debian:trixie-slim

WORKDIR /usr/src/app

COPY --from=builder /usr/local/bin/app /usr/local/bin/app
COPY --from=builder /usr/src/app/templates ./templates

EXPOSE 80
CMD ["/usr/local/bin/app"]
