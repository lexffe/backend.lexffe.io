# linux/amd64 musl-libc

FROM golang:alpine as compile

RUN mkdir /app
ADD . /app
WORKDIR /app

RUN apk add --no-cache build-base gcc musl-dev

RUN go mod download

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -ldflags='-s -w -linkmode external -extldflags "-fno-PIC -static"' -buildmode pie -tags 'osusergo netgo static_build' -o main .

FROM busybox:musl

RUN mkdir /app
WORKDIR /app
COPY --from=compile /app/main .

EXPOSE 8080

CMD ["/app/main"]

HEALTHCHECK --interval=10m --timeout=30s --start-period=5s --retries=3 CMD [ "curl -f 127.0.0.1:8080" ]