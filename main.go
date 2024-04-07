package main

import (
	"flag"
	"io"
	"log"
	"net/http"
)

// Define a command-line flag
var debugMode = flag.Bool("debug", false, "Enable debug mode")

func handleRequestAndRedirect(w http.ResponseWriter, r *http.Request) {
	if *debugMode {
		log.Printf("Received request: %s %s", r.Method, r.URL)
	}

	outReq := r.Clone(r.Context())

	resp, err := http.DefaultTransport.RoundTrip(outReq)
	if err != nil {
		http.Error(w, "Error forwarding request: "+err.Error(), http.StatusInternalServerError)
		log.Printf("Error forwarding request: %v", err)
		return
	}
	defer resp.Body.Close()

	if *debugMode {
		log.Printf("Received response: %d for %s", resp.StatusCode, r.URL)
	}

	copyHeader(w.Header(), resp.Header)

	w.WriteHeader(resp.StatusCode)

	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("Error copying response body: %v", err)
	}
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func main() {
	// Parse the command-line flags
	flag.Parse()
	if *debugMode {
		log.Println("Debug mode is enabled")
	}
	http.HandleFunc("/", handleRequestAndRedirect)
	log.Println("Server started on :8510")
	log.Fatal(http.ListenAndServe(":8510", nil))
}
