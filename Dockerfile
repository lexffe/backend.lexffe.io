FROM golang:alpine

RUN mkdir /app
ADD . /app
WORKDIR /app

RUN go mod download

RUN go build -o main .

EXPOSE 8080

CMD ["/app/main"]

HEALTHCHECK --interval=10m --timeout=30s --start-period=5s --retries=3 CMD [ "curl -f 127.0.0.1:8080" ]