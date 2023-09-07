FROM golang:1.21

WORKDIR /usr/src/app
COPY . .

RUN go build -o app . && mv app /usr/local/bin/app

ENV GIN_MODE release
EXPOSE 8080
CMD ["app"]
