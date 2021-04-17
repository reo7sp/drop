FROM golang:1.14

WORKDIR /go/src/app
COPY . .

RUN go get -d ./...
RUN go install ./...

ENV GIN_MODE release
EXPOSE 8080
CMD ["app"]
