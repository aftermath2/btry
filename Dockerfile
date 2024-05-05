FROM golang:1.22-alpine3.19 as builder

RUN apk add --update --no-cache git nodejs npm ca-certificates && update-ca-certificates

WORKDIR /app

COPY go.mod .

RUN go mod download

COPY . .

RUN npm ci --prefix ui/

RUN npm run build --prefix ui/

RUN CGO_ENABLED=0 go build -o BTRY -ldflags="-s -w" .

# -----

FROM scratch

COPY --from=builder /app/BTRY /usr/bin/BTRY

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/usr/bin/BTRY"]
