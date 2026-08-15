FROM golang:1.26.6

WORKDIR /usr/src/app
COPY . .

RUN go build -o app . && mv app /usr/local/bin/app

EXPOSE 8080
CMD ["app"]
