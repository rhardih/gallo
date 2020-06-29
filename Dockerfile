FROM golang:1.14-alpine AS build

RUN apk update && apk add --no-cache git

WORKDIR /gallo

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN ./scripts/hash_name_assets.sh

RUN CGO_ENABLED=0 GOOS=linux \
  go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main .

FROM scratch

WORKDIR /gallo

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /gallo/public/assets/css ./public/assets/css
COPY --from=build /gallo/public/assets/icons ./public/assets/icons
COPY --from=build /gallo/public/assets/images ./public/assets/images
COPY --from=build /gallo/public/assets/js ./public/assets/js
COPY --from=build /gallo/public/assets/vendor ./public/assets/vendor
COPY --from=build /gallo/public/assets/videos ./public/assets/videos
COPY --from=build /gallo/app/views ./app/views
COPY --from=build /gallo/main ./main
COPY --from=build /gallo/public ./public

EXPOSE 8080

CMD ["./main"]
