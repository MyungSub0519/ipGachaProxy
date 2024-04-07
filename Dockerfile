FROM golang:1.22.1 as builder

WORKDIR /app

ENV GO111MODULE=on

COPY go.mod ./

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ipgachaproxy .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

EXPOSE 8080

COPY --from=builder /app/ipgachaproxy .

CMD ["./ipgachaproxy"]