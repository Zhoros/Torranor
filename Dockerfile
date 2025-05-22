FROM alpine:latest

RUN apk add --no-cache git go g++

WORKDIR /app

COPY . /app

RUN go build .

CMD ["./app"]
