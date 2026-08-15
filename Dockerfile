FROM golang:1.26.6

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /usr/local/bin/app .

EXPOSE 8080
CMD ["app"]
