FROM golang:1.15-alpine

RUN apk update && apk add --no-cache \
  gcc \
  git \
  musl-dev

WORKDIR /gallo

# Do a trial compile here instead of waiting for it to fail during at run time
COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build .

# Build succeeded, so it's fair to assume it'll do the same with CompileDaemon
RUN go get github.com/githubnemo/CompileDaemon

EXPOSE 8080

CMD ["CompileDaemon", \
  "-color=true", \
  "-exclude-dir=.git", \
  "-graceful-kill=true", \
  "-graceful-timeout=10", \
  "-command=./gallo"]
