package main

import (
	"flag"
	"io"
	"log"
	"net"
	"net/http"
)

var debugMode = flag.Bool("debug", false, "Enable debug mode")

func handleHTTPS(w http.ResponseWriter, r *http.Request) {
    if *debugMode {
        log.Printf("Received HTTPS request for: %s", r.Host)
    }

    destConn, err := net.Dial("tcp", r.Host)
    if err != nil {
        http.Error(w, "Failed to connect to host", http.StatusInternalServerError)
        return
    }
    defer destConn.Close()

    w.WriteHeader(http.StatusOK)

    hijacker, ok := w.(http.Hijacker)
    if !ok {
        http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
        return
    }

    clientConn, _, err := hijacker.Hijack()
    if err != nil {
        http.Error(w, "Hijacking failed", http.StatusInternalServerError)
        return
    }
    defer clientConn.Close()

    go transfer(destConn, clientConn)
    go transfer(clientConn, destConn)
}


func handleHTTP(w http.ResponseWriter, r *http.Request) {
	if *debugMode {
		log.Printf("Received request: %s %s", r.Method, r.URL)
	}

	outReq := r.Clone(r.Context())

	resp, err := http.DefaultTransport.RoundTrip(outReq)
	if err != nil {
		http.Error(w, "Error forwarding request: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if *debugMode {
		log.Printf("Received response: %d for %s", resp.StatusCode, r.URL)
	}

	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		handleHTTPS(w, r)
	} else {
		handleHTTP(w, r)
	}
}

func main() {
	flag.Parse()

	if *debugMode {
		log.Println("Debug mode is enabled")
	}

	http.HandleFunc("/", handleRequest)
	log.Println("Proxy server started on :8510")
	log.Fatal(http.ListenAndServe(":8510", nil))
}
