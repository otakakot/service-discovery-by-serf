FROM golang:1.23.4-alpine

WORKDIR /app

RUN go install github.com/cosmtrek/air@latest

WORKDIR /app

CMD ["air"]
