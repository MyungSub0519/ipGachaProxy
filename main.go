package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

func main() {
	// TCP Listen on host:port
	const host = "0.0.0.0"
	const port = "8510"
	bindingAddress := fmt.Sprintf("%s:%s", host, port)
	listener, err := net.Listen("tcp", bindingAddress)
	if err != nil {
		panic(err)
	}
	log.Printf("ipGachaProxy is listen on %s", bindingAddress)
	for {
		connection, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		go handleConnection(connection.(*net.TCPConn))
	}
}

func handleConnection(clientConn *net.TCPConn) {
	defer clientConn.Close()

	clientIP := clientConn.RemoteAddr().(*net.TCPAddr).IP
	connectionID := uuid.New().String()
	reader := bufio.NewReader(clientConn)

	peekBytes, err := reader.Peek(3)
	if err != nil {
		log.Println("Error peeking at connection bytes:", err)
		// TODO : Solving error reader.peek(3) return EOF err
		// log.Println(reader)
		return
	}

	if string(peekBytes) == "GET" || string(peekBytes) == "POS" {
		handleHTTPRequest(reader, clientConn, clientIP, connectionID)
		return
	}

	serverNameIndication := getClientHelloServerName(reader)
	if serverNameIndication == "" {
		return
	}

	hostAddresses, err := net.LookupHost(serverNameIndication)
	if err != nil || len(hostAddresses) < 1 {
		logConnectionError(clientIP, connectionID, serverNameIndication, err)
		return
	}

	serverConn, err := net.DialTimeout("tcp", hostAddresses[0]+":443", 10*time.Second)
	if err != nil {
		log.Printf("%s SNI_TRANSPARENT/500 CONNECT %s %s - CONN_FAIL/%s", clientIP, connectionID, serverNameIndication+":443", hostAddresses[0])
		return
	}
	defer serverConn.Close()

	clientWrappedConn := wrapConnection(reader, clientConn)
	serverWrappedConn := &Conn{Conn: serverConn}

	var wg sync.WaitGroup
	wg.Add(2)
	go proxy(&wg, clientWrappedConn, serverWrappedConn)
	go proxy(&wg, serverWrappedConn, clientWrappedConn)
	log.Printf("%s SNI_TRANSPARENT/200 CONNECT %s %s - SNI_DIRECT/%s", clientIP, connectionID, serverNameIndication+":443", hostAddresses[0])
	wg.Wait()
	log.Printf("%s SNI_TRANSPARENT/200 CLOSE   %s %s - SNI_DIRECT/%s", clientIP, connectionID, serverNameIndication+":443", hostAddresses[0])
}

func handleHTTPRequest(reader *bufio.Reader, clientConn net.Conn, clientIP net.IP, connectionID string) {
	request, err := http.ReadRequest(reader)
	if err != nil {
		log.Printf("Error reading HTTP request: %v", err)
		return
	}

	log.Printf("%s HTTP REQUEST %s %s %s", clientIP, connectionID, request.Method, request.RequestURI)

	response, err := forwardHTTPRequest(request)
	if err != nil {
		log.Printf("Error forwarding HTTP request: %v", err)
		return
	}

	err = response.Write(clientConn)
	if err != nil {
		log.Printf("Error writing HTTP response to client: %v", err)
	}
}

func forwardHTTPRequest(req *http.Request) (*http.Response, error) {
	url := req.URL.String()
	if !req.URL.IsAbs() {
		url = "http://" + req.Host + req.URL.String()
	}

	forwardReq, err := http.NewRequest(req.Method, url, req.Body)
	if err != nil {
		return nil, err
	}
	forwardReq.Header = req.Header

	client := &http.Client{}
	resp, err := client.Do(forwardReq)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type Conn struct {
	Peeked []byte
	net.Conn
}

func (c *Conn) Read(p []byte) (n int, err error) {
	if len(c.Peeked) > 0 {
		n = copy(p, c.Peeked)
		c.Peeked = c.Peeked[n:]
		if len(c.Peeked) == 0 {
			c.Peeked = nil
		}
		return n, nil
	}
	return c.Conn.Read(p)
}

func proxy(wg *sync.WaitGroup, source, destination *Conn) {
	defer wg.Done()
	if _, err := io.Copy(source.Conn.(*net.TCPConn), destination); err != nil {
	}
	source.Conn.(*net.TCPConn).CloseRead()
	destination.Conn.(*net.TCPConn).CloseWrite()
}

func getClientHelloServerName(reader *bufio.Reader) string {
	const recordHeaderLen = 5
	header, err := reader.Peek(recordHeaderLen)
	if err != nil || header[0] != 0x16 {
		return ""
	}

	recordLength := int(header[3])<<8 | int(header[4])
	helloBytes, err := reader.Peek(recordHeaderLen + recordLength)
	if err != nil {
		return ""
	}

	var serverNameIndication string
	tls.Server(sniSniffConn{reader: bytes.NewReader(helloBytes)}, &tls.Config{
		GetConfigForClient: func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
			serverNameIndication = hello.ServerName
			return nil, nil
		},
	}).Handshake()
	return serverNameIndication
}

type sniSniffConn struct {
	reader io.Reader
	net.Conn
}

func (conn sniSniffConn) Read(p []byte) (int, error) { return conn.reader.Read(p) }
func (sniSniffConn) Write(p []byte) (int, error)     { return 0, io.EOF }

func wrapConnection(reader *bufio.Reader, conn net.Conn) *Conn {
	if n := reader.Buffered(); n > 0 {
		peeked, _ := reader.Peek(reader.Buffered())
		return &Conn{Peeked: peeked, Conn: conn}
	}
	return &Conn{Conn: conn}
}

func logConnectionError(clientIP net.IP, connectionID, serverNameIndication string, err error) {
	if err != nil {
		log.Printf("%s SNI_TRANSPARENT/500 CONNECT %s %s - DNS_FAIL/%s", clientIP, connectionID, serverNameIndication+":443", err)
	} else {
		log.Printf("%s SNI_TRANSPARENT/500 CONNECT %s %s - DNS_FAIL/NONE", clientIP, connectionID, serverNameIndication+":443")
	}
}
