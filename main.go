package main

import (
	"flag"
	"log"
	"net/http"

	"sepolia-bundle-api/server"
)

func main() {
	port := flag.String("port", "8080", "Server port")
	flag.Parse()

	// Create new server instance
	s := server.NewRpcEndPointServer()

	// Setup HTTP routes
	http.HandleFunc("/bundle", s.HandleBundleRequest)

	// Start server
	log.Printf("Starting server on port %s", *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatal(err)
	}
}
