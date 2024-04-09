FROM golang:1.22.1 as builder

WORKDIR /app

ENV GO111MODULE=on

COPY go.mod go.sum ./

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ipgachaproxy .

FROM alpine:latest

WORKDIR /root/

EXPOSE 8510

COPY --from=builder /app/ipgachaproxy .

CMD ["./ipgachaproxy"]