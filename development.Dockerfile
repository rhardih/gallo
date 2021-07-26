FROM golang:1.14-alpine

RUN apk update && apk add --no-cache \
  gcc \
  git \
  musl-dev

WORKDIR /gallo

RUN go get github.com/githubnemo/CompileDaemon

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Build here to trigger build errors early
RUN go build -o main .

EXPOSE 8080
