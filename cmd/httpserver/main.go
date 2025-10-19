package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"boot.httpserver/internal/server"
)

const port = 42069

func main() {
	srv, err := server.Serve(port)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer srv.Close()
	log.Printf("Server started on port: %d\n", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Printf("Server gracefully stopped\n")
}
