FROM golang:1.14-alpine

RUN apk update && apk add --no-cache \
  gcc \
  git \
  musl-dev

WORKDIR /gallo

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main .

EXPOSE 8080
