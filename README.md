```
go mod init github.com/MyungSub0519/ipGachaProxy
go mod tidy
go build -a -installsuffix cgo -o ipgachaproxy .
./ipgachaproxy
```


```
docker build . -t ipgachaproxy:0.0.1
docker run ipgachaproxy:0.0.1 -p 8510:8510
```