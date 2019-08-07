FROM golang:1.12-alpine as builder

RUN mkdir /build

COPY . /build/go-whosonfirst-rasterzen

RUN apk update && apk upgrade \
    && apk add make libc-dev gcc \
    && cd /build/go-whosonfirst-rasterzen \
    && make tools

FROM alpine:latest

COPY --from=builder /build/go-whosonfirst-rasterzen/bin/wof-rasterzen-seed /usr/local/bin/wof-rasterzen-seed

RUN apk update && apk upgrade \
    && apk add ca-certificates