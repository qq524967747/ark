FROM golang:1.18 as build

WORKDIR /build

COPY . /build


RUN export GO111MODULE=on && \
    export CGO_ENABLED=0 && \
    export GOPROXY=https://goproxy.cn && \
    go build -o qq-bot .



FROM alpine

COPY --from=build /build/device.json /usr/local/bin
COPY --from=build /build/mirai.yml  /usr/local/bin
COPY --from=build /build/session.token  /usr/local/bin
COPY --from=build /build/qq-bot  /usr/local/bin

RUN chmod +x /usr/local/bin/qq-bot

ENTRYPOINT qq-bot
