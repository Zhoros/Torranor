FROM golang:1.24-alpine

WORKDIR /app

COPY . /app

RUN go build -o /bin/app .

FROM scratch

COPY --from=0 /bin/app /app
COPY  ./public /public
CMD ["./app"]