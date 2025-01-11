# syntax=docker/dockerfile:1

FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY cmd ./cmd
COPY pkg ./pkg

RUN go build -o /bot cmd/main/main.go

RUN pwd

CMD [ "/bot" ]
