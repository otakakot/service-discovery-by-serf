FROM golang:1.22.3-alpine

WORKDIR /app

RUN go install github.com/cosmtrek/air@latest

WORKDIR /app

CMD ["air"]
