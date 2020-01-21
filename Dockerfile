FROM golang:1.12.0-alpine3.9

RUN apk update && apk add alpine-sdk git && rm -rf /var/cache/apk/*

RUN mkdir -p /api
WORKDIR /api

RUN go get -u github.com/gin-gonic/gin
RUN go get -u github.com/gocolly/colly/
RUN go get -u github.com/chilts/sid

COPY . .
RUN go build -o ./app ./src

ENTRYPOINT ["./app"]
