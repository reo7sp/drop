FROM golang:1.21

WORKDIR /go/src/app
COPY . .

RUN go build -o app .

ENV GIN_MODE release
EXPOSE 8080
CMD ["app"]
