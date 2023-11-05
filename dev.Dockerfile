FROM golang:1.21-alpine3.18

RUN apk add --update --no-cache ca-certificates && update-ca-certificates

WORKDIR /BTRY

COPY go.mod .

RUN go mod download

CMD ["/bin/sh"]