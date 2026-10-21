FROM golang:1.25.0-alpine3.22 AS build-dev

WORKDIR /go/src/app

COPY . .

RUN go env -w GO111MODULE=on && \
    go env -w GOPROXY=https://goproxy.cn,direct && \
    go mod download && \
    apk add --no-cache upx && \
    go build -trimpath  -ldflags="-s -w" -o short . && \
    [ -e /usr/bin/upx ] && upx short || echo

FROM alpine

RUN apk update && \
    apk add --no-cache tzdata && \
    rm -rf /var/cache/apk/*

ENV TZ=Asia/Shanghai

COPY --from=build-dev /go/src/app/short ./short

CMD ["./short"]