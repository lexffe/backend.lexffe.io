# linux/amd64 musl-libc

FROM golang:alpine as compile

RUN mkdir /app
ADD . /app
WORKDIR /app

RUN go mod download

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o main .

FROM alpine:latest as packer

RUN apk add --no-cache upx
RUN mkdir /app
WORKDIR /app

COPY --from=compile /app/main .
COPY --from=compile /app/config.toml .

RUN upx --ultra-brute --best main

FROM busybox:musl

RUN mkdir /app
WORKDIR /app
COPY --from=packer /app/main .
COPY --from=compile /app/config.toml .

EXPOSE 8080

CMD ["/app/main"]

HEALTHCHECK --interval=10m --timeout=30s --start-period=5s --retries=3 CMD [ "curl -f 127.0.0.1:8080" ]