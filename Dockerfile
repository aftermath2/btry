FROM golang:1.22-alpine3.18 as builder

RUN apk add --update --no-cache ca-certificates && update-ca-certificates

WORKDIR /app

COPY go.mod .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o BTRY -ldflags="-s -w" ./main.go

# -----

FROM scratch

COPY --from=builder /app/BTRY /usr/bin/BTRY

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/usr/bin/BTRY"]
