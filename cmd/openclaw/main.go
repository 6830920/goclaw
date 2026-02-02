package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

// OpenClaw-Go - A personal AI assistant implemented in Go
//
// This is a reimplementation of the original OpenClaw project (https://github.com/openclaw/openclaw)
// which provides a personal AI assistant that can integrate with various communication channels.

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "18789" // Default OpenClaw port
	}
	
	fmt.Printf("Starting OpenClaw-Go on port %s\n", port)
	
	http.HandleFunc("/", handleRoot)
	
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OpenClaw-Go - Personal AI Assistant\n")
	fmt.Fprintf(w, "Gateway server running\n")
}